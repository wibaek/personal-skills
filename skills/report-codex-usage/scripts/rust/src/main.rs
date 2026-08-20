use std::collections::{BTreeMap, HashMap, HashSet};
use std::env;
use std::fs::File;
use std::io::{BufRead, BufReader};
use std::path::{Path, PathBuf};
use std::process::{Command, ExitCode};

use anyhow::{Context, Result, bail};
use chrono::{DateTime, Datelike, NaiveDate, SecondsFormat, TimeZone, Utc};
use chrono_tz::Tz;
use clap::{Parser, ValueEnum};
use rayon::prelude::*;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use walkdir::WalkDir;

const MODEL_LABEL_ORDER: [&str; 7] = ["sol", "terra", "luna", "5.5", "5.4", "review", "other"];

#[derive(Parser)]
#[command(about = "Aggregate daily Codex token usage from local session JSONL metadata.")]
struct Cli {
    #[arg(
        long,
        default_value = "yesterday",
        help = "today, yesterday, or YYYY-MM-DD (default: yesterday)"
    )]
    date: String,

    #[arg(
        long,
        default_value = "Asia/Seoul",
        help = "IANA timezone used for the calendar-day boundary"
    )]
    timezone: String,

    #[arg(
        long,
        help = "override the default active and archived Codex session directories"
    )]
    sessions_root: Option<PathBuf>,

    #[arg(long, help = "TOML rate card")]
    rate_card: Option<PathBuf>,

    #[arg(long, help = "Codex session title index")]
    session_index: Option<PathBuf>,

    #[arg(
        long,
        help = "Codex desktop global state containing UI project assignments"
    )]
    global_state: Option<PathBuf>,

    #[arg(long, help = "Override the detected computer name")]
    computer_name: Option<String>,

    #[arg(long, value_enum, default_value_t = OutputFormat::Markdown)]
    format: OutputFormat,
}

#[derive(Clone, Copy, ValueEnum)]
enum OutputFormat {
    Markdown,
    Json,
}

#[derive(Clone, Deserialize)]
struct Rate {
    model: String,
    #[serde(deserialize_with = "deserialize_date")]
    effective_from: NaiveDate,
    input_per_million: f64,
    cached_input_per_million: f64,
    output_per_million: f64,
    cache_write_per_million: f64,
}

#[derive(Deserialize)]
struct RateCard {
    #[serde(default)]
    rate: Vec<Rate>,
}

#[derive(Clone, Default, Serialize)]
struct Totals {
    total: u64,
    cached_input: u64,
    input: u64,
    output: u64,
    calculated_cost: f64,
    events: u64,
}

impl Totals {
    fn from_usage(usage: &Value, rate: Option<&Rate>) -> Result<Self, String> {
        let usage = usage
            .as_object()
            .ok_or_else(|| "last_token_usage must be an object".to_string())?;
        let input_tokens = integer_token(usage.get("input_tokens"), "input_tokens")?;
        let cached_input = integer_token(usage.get("cached_input_tokens"), "cached_input_tokens")?;
        let cache_write = integer_token(
            usage.get("cache_write_input_tokens"),
            "cache_write_input_tokens",
        )?;
        let output = integer_token(usage.get("output_tokens"), "output_tokens")?;

        if cached_input
            .checked_add(cache_write)
            .is_none_or(|value| value > input_tokens)
        {
            return Err(
                "cached_input_tokens + cache_write_input_tokens exceeds input_tokens".to_string(),
            );
        }

        let mut totals = Self {
            total: input_tokens + output,
            cached_input,
            input: input_tokens - cached_input,
            output,
            calculated_cost: 0.0,
            events: 1,
        };

        if let Some(rate) = rate {
            let regular_input = input_tokens - cached_input - cache_write;
            totals.calculated_cost = (regular_input as f64 * rate.input_per_million
                + cached_input as f64 * rate.cached_input_per_million
                + cache_write as f64 * rate.cache_write_per_million
                + output as f64 * rate.output_per_million)
                / 1_000_000.0;
        }

        Ok(totals)
    }

    fn merge(&mut self, other: &Self) {
        self.total += other.total;
        self.cached_input += other.cached_input;
        self.input += other.input;
        self.output += other.output;
        self.calculated_cost += other.calculated_cost;
        self.events += other.events;
    }
}

#[derive(Clone, Default, Serialize)]
struct Diagnostics {
    files_scanned: u64,
    files_with_target_events: u64,
    original_events: u64,
    duplicate_events: u64,
    replayed_events: u64,
    aggregated_events: u64,
    malformed_lines: u64,
    unreadable_files: u64,
    token_events_without_usage: u64,
    invalid_token_events: u64,
}

impl Diagnostics {
    fn merge(&mut self, other: &Self) {
        self.files_scanned += other.files_scanned;
        self.files_with_target_events += other.files_with_target_events;
        self.original_events += other.original_events;
        self.duplicate_events += other.duplicate_events;
        self.replayed_events += other.replayed_events;
        self.aggregated_events += other.aggregated_events;
        self.malformed_lines += other.malformed_lines;
        self.unreadable_files += other.unreadable_files;
        self.token_events_without_usage += other.token_events_without_usage;
        self.invalid_token_events += other.invalid_token_events;
    }
}

struct SessionUsage {
    session_id: String,
    title: String,
    models: HashSet<String>,
    totals: Totals,
}

struct Report {
    target_date: NaiveDate,
    timezone_name: String,
    range_start: DateTime<Tz>,
    range_end: DateTime<Tz>,
    generated_at: DateTime<Tz>,
    computer_name: String,
    projects: HashMap<String, Totals>,
    models: HashMap<String, Totals>,
    sessions: HashMap<String, HashMap<String, SessionUsage>>,
    total: Totals,
    diagnostics: Diagnostics,
}

#[derive(Default, Deserialize)]
struct RawRow {
    #[serde(rename = "type")]
    row_type: Option<String>,
    timestamp: Option<String>,
    payload: Option<RawPayload>,
}

#[derive(Default, Deserialize)]
struct RawPayload {
    id: Option<Value>,
    session_id: Option<Value>,
    model: Option<Value>,
    #[serde(rename = "type")]
    payload_type: Option<Value>,
    thread_settings: Option<Value>,
    info: Option<Value>,
}

enum Record {
    SessionMeta {
        id: Option<String>,
        session_id: Option<String>,
    },
    Model(Option<String>),
    TaskStarted,
    TokenCount {
        timestamp: Option<DateTime<Utc>>,
        info: Option<Value>,
    },
}

struct Candidate {
    identity: (String, String, String, String),
    model: String,
    usage: Value,
}

struct FileOutcome {
    report_session_id: String,
    candidates: Vec<Candidate>,
    diagnostics: Diagnostics,
}

fn deserialize_date<'de, D>(deserializer: D) -> Result<NaiveDate, D::Error>
where
    D: serde::Deserializer<'de>,
{
    let raw = String::deserialize(deserializer)?;
    NaiveDate::parse_from_str(&raw, "%Y-%m-%d").map_err(serde::de::Error::custom)
}

fn integer_token(value: Option<&Value>, key: &str) -> Result<u64, String> {
    match value {
        None | Some(Value::Null) => Ok(0),
        Some(Value::Number(number)) => number
            .as_u64()
            .ok_or_else(|| format!("{key} must be a non-negative integer")),
        _ => Err(format!("{key} must be a non-negative integer")),
    }
}

fn value_string(value: Option<&Value>) -> Option<String> {
    value
        .and_then(Value::as_str)
        .filter(|value| !value.is_empty())
        .map(str::to_owned)
}

fn parse_timestamp(value: Option<&str>) -> Option<DateTime<Utc>> {
    let value = value?;
    DateTime::parse_from_rfc3339(value)
        .map(|timestamp| timestamp.with_timezone(&Utc))
        .or_else(|_| {
            chrono::NaiveDateTime::parse_from_str(value, "%Y-%m-%dT%H:%M:%S%.f")
                .map(|timestamp| timestamp.and_utc())
        })
        .ok()
}

fn load_rates(path: &Path) -> Result<HashMap<String, Vec<Rate>>> {
    let raw = std::fs::read_to_string(path)
        .with_context(|| format!("rate card를 읽을 수 없음: {}", path.display()))?;
    let mut card: RateCard = toml::from_str(&raw)
        .with_context(|| format!("rate card 형식이 잘못됨: {}", path.display()))?;
    let mut rates: HashMap<String, Vec<Rate>> = HashMap::new();
    for rate in card.rate.drain(..) {
        rates.entry(rate.model.clone()).or_default().push(rate);
    }
    for model_rates in rates.values_mut() {
        model_rates.sort_by_key(|rate| rate.effective_from);
    }
    Ok(rates)
}

fn rate_for<'a>(
    rates: &'a HashMap<String, Vec<Rate>>,
    model: &str,
    target_date: NaiveDate,
) -> Option<&'a Rate> {
    rates
        .get(model)?
        .iter()
        .rev()
        .find(|rate| rate.effective_from <= target_date)
}

#[derive(Deserialize)]
struct SessionIndexRow {
    id: Option<String>,
    thread_name: Option<String>,
}

fn load_session_titles(path: &Path) -> HashMap<String, String> {
    let Ok(file) = File::open(path) else {
        return HashMap::new();
    };
    let mut titles = HashMap::new();
    let mut reader = BufReader::new(file);
    let mut line = Vec::new();
    loop {
        line.clear();
        let Ok(bytes_read) = reader.read_until(b'\n', &mut line) else {
            break;
        };
        if bytes_read == 0 {
            break;
        }
        let Ok(row) = serde_json::from_slice::<SessionIndexRow>(&line) else {
            continue;
        };
        if let (Some(session_id), Some(title)) = (row.id, row.thread_name)
            && !title.is_empty()
        {
            titles.insert(session_id, title);
        }
    }
    titles
}

#[derive(Deserialize)]
struct GlobalState {
    #[serde(rename = "local-projects", default)]
    local_projects: HashMap<String, LocalProject>,
    #[serde(rename = "thread-project-assignments", default)]
    thread_project_assignments: HashMap<String, ProjectAssignment>,
}

#[derive(Deserialize)]
struct LocalProject {
    name: Option<String>,
}

#[derive(Deserialize)]
struct ProjectAssignment {
    #[serde(rename = "projectKind")]
    project_kind: Option<String>,
    #[serde(rename = "projectId")]
    project_id: Option<String>,
}

fn load_project_assignments(path: &Path) -> HashMap<String, String> {
    let Ok(file) = File::open(path) else {
        return HashMap::new();
    };
    let Ok(state) = serde_json::from_reader::<_, GlobalState>(BufReader::new(file)) else {
        return HashMap::new();
    };

    let project_names: HashMap<String, String> = state
        .local_projects
        .into_iter()
        .filter_map(|(project_id, project)| {
            project
                .name
                .filter(|name| !name.is_empty())
                .map(|name| (project_id, name))
        })
        .collect();

    state
        .thread_project_assignments
        .into_iter()
        .filter_map(|(thread_id, assignment)| {
            if assignment.project_kind.as_deref() != Some("local") {
                return None;
            }
            let project_id = assignment.project_id?;
            project_names
                .get(&project_id)
                .cloned()
                .map(|name| (thread_id, name))
        })
        .collect()
}

fn model_label(model: &str) -> &'static str {
    let normalized = model.to_lowercase();
    if matches!(normalized.as_str(), "gpt-5.6" | "gpt-5.6-sol") || normalized.contains("5.6-sol") {
        "sol"
    } else if normalized.contains("terra") {
        "terra"
    } else if normalized.contains("luna") {
        "luna"
    } else if normalized == "gpt-5.5" || normalized.starts_with("gpt-5.5-") {
        "5.5"
    } else if normalized == "gpt-5.4" || normalized.starts_with("gpt-5.4-") {
        "5.4"
    } else if normalized == "codex-auto-review" || normalized.contains("auto-review") {
        "review"
    } else {
        "other"
    }
}

fn process_file(path: &Path, range_start: DateTime<Utc>, range_end: DateTime<Utc>) -> FileOutcome {
    let mut diagnostics = Diagnostics {
        files_scanned: 1,
        ..Diagnostics::default()
    };
    let Ok(file) = File::open(path) else {
        diagnostics.unreadable_files += 1;
        return FileOutcome {
            report_session_id: path_stem(path),
            candidates: Vec::new(),
            diagnostics,
        };
    };

    let mut records = Vec::new();
    let mut file_models = HashSet::new();
    let mut reader = BufReader::with_capacity(1024 * 1024, file);
    let mut line = Vec::new();
    loop {
        line.clear();
        let bytes_read = match reader.read_until(b'\n', &mut line) {
            Ok(bytes_read) => bytes_read,
            Err(_) => {
                diagnostics.unreadable_files += 1;
                break;
            }
        };
        if bytes_read == 0 {
            break;
        }

        let row = match serde_json::from_slice::<RawRow>(&line) {
            Ok(row) => row,
            Err(_) => {
                diagnostics.malformed_lines += 1;
                continue;
            }
        };
        let payload = row.payload.unwrap_or_default();
        match row.row_type.as_deref() {
            Some("session_meta") => records.push(Record::SessionMeta {
                id: value_string(payload.id.as_ref()),
                session_id: value_string(payload.session_id.as_ref()),
            }),
            Some("turn_context") => {
                let model = value_string(payload.model.as_ref());
                if let Some(model) = &model {
                    file_models.insert(model.clone());
                }
                records.push(Record::Model(model));
            }
            Some("event_msg") => match value_string(payload.payload_type.as_ref()).as_deref() {
                Some("thread_settings_applied") => {
                    let model = payload
                        .thread_settings
                        .as_ref()
                        .and_then(Value::as_object)
                        .and_then(|settings| value_string(settings.get("model")));
                    if let Some(model) = &model {
                        file_models.insert(model.clone());
                    }
                    records.push(Record::Model(model));
                }
                Some("task_started") => records.push(Record::TaskStarted),
                Some("token_count") => records.push(Record::TokenCount {
                    timestamp: parse_timestamp(row.timestamp.as_deref()),
                    info: payload.info,
                }),
                _ => {}
            },
            _ => {}
        }
    }

    let first_meta = records.iter().find_map(|record| match record {
        Record::SessionMeta { id, session_id } => Some((id.clone(), session_id.clone())),
        _ => None,
    });
    let fallback_id = path_stem(path);
    let rollout_id = first_meta
        .as_ref()
        .and_then(|(id, _)| id.clone())
        .unwrap_or_else(|| fallback_id.clone());
    let report_session_id = first_meta
        .and_then(|(_, session_id)| session_id)
        .unwrap_or_else(|| rollout_id.clone());

    let last_foreign_meta =
        records
            .iter()
            .enumerate()
            .rev()
            .find_map(|(index, record)| match record {
                Record::SessionMeta { id, .. } if id.as_deref() != Some(rollout_id.as_str()) => {
                    Some(index)
                }
                _ => None,
            });
    let replay_cutoff = last_foreign_meta.and_then(|foreign_index| {
        records
            .iter()
            .enumerate()
            .skip(foreign_index + 1)
            .find_map(|(index, record)| matches!(record, Record::TaskStarted).then_some(index))
    });
    let replay_only_rollout = last_foreign_meta.is_some() && replay_cutoff.is_none();
    let unique_file_model = (file_models.len() == 1)
        .then(|| file_models.iter().next().cloned())
        .flatten();

    let mut current_model: Option<String> = None;
    let mut candidates = Vec::new();
    let mut file_has_target_event = false;
    for (record_index, record) in records.into_iter().enumerate() {
        match record {
            Record::Model(model) => {
                if model.is_some() {
                    current_model = model;
                }
            }
            Record::TokenCount { timestamp, info } => {
                let Some(timestamp) = timestamp else {
                    continue;
                };
                if timestamp < range_start || timestamp >= range_end {
                    continue;
                }

                diagnostics.original_events += 1;
                file_has_target_event = true;
                if replay_only_rollout || replay_cutoff.is_some_and(|cutoff| record_index < cutoff)
                {
                    diagnostics.replayed_events += 1;
                    continue;
                }

                let Some(info) = info.and_then(|value| value.as_object().cloned()) else {
                    diagnostics.token_events_without_usage += 1;
                    continue;
                };
                let Some(usage) = info
                    .get("last_token_usage")
                    .filter(|value| value.is_object())
                    .cloned()
                else {
                    diagnostics.token_events_without_usage += 1;
                    continue;
                };
                let total_usage = info
                    .get("total_token_usage")
                    .cloned()
                    .unwrap_or(Value::Null);
                let timestamp_key = timestamp.to_rfc3339_opts(SecondsFormat::AutoSi, false);
                let total_key = serde_json::to_string(&total_usage).unwrap_or_default();
                let usage_key = serde_json::to_string(&usage).unwrap_or_default();
                let model = current_model
                    .clone()
                    .or_else(|| unique_file_model.clone())
                    .unwrap_or_else(|| "미분류".to_string());
                candidates.push(Candidate {
                    identity: (rollout_id.clone(), timestamp_key, total_key, usage_key),
                    model,
                    usage,
                });
            }
            Record::SessionMeta { .. } | Record::TaskStarted => {}
        }
    }
    if file_has_target_event {
        diagnostics.files_with_target_events += 1;
    }

    FileOutcome {
        report_session_id,
        candidates,
        diagnostics,
    }
}

fn path_stem(path: &Path) -> String {
    path.file_stem()
        .and_then(|value| value.to_str())
        .unwrap_or("unknown-rollout")
        .to_string()
}

fn collect_session_files(roots: &[PathBuf]) -> Vec<PathBuf> {
    let mut files: Vec<PathBuf> = roots
        .iter()
        .flat_map(|root| {
            WalkDir::new(root)
                .follow_links(false)
                .into_iter()
                .filter_map(Result::ok)
                .filter(|entry| entry.file_type().is_file())
                .filter(|entry| entry.path().extension().is_some_and(|ext| ext == "jsonl"))
                .map(|entry| entry.into_path())
        })
        .collect();
    files.sort();
    files
}

#[allow(clippy::too_many_arguments)]
fn aggregate(
    roots: &[PathBuf],
    target_date: NaiveDate,
    timezone: Tz,
    timezone_name: String,
    rate_card: &Path,
    session_index: &Path,
    global_state: &Path,
    computer_name: String,
) -> Result<Report> {
    let range_start = timezone
        .with_ymd_and_hms(
            target_date.year(),
            target_date.month(),
            target_date.day(),
            0,
            0,
            0,
        )
        .single()
        .context("집계 시작 시각을 만들 수 없음")?;
    let next_date = target_date.succ_opt().context("집계 종료 날짜 범위 초과")?;
    let range_end = timezone
        .with_ymd_and_hms(
            next_date.year(),
            next_date.month(),
            next_date.day(),
            0,
            0,
            0,
        )
        .single()
        .context("집계 종료 시각을 만들 수 없음")?;
    let range_start_utc = range_start.with_timezone(&Utc);
    let range_end_utc = range_end.with_timezone(&Utc);

    let rates = load_rates(rate_card)?;
    let session_titles = load_session_titles(session_index);
    let project_assignments = load_project_assignments(global_state);
    let files = collect_session_files(roots);
    let outcomes: Vec<FileOutcome> = files
        .par_iter()
        .map(|path| process_file(path, range_start_utc, range_end_utc))
        .collect();

    let mut projects: HashMap<String, Totals> = HashMap::new();
    let mut models: HashMap<String, Totals> = HashMap::new();
    let mut sessions: HashMap<String, HashMap<String, SessionUsage>> = HashMap::new();
    let mut total = Totals::default();
    let mut diagnostics = Diagnostics::default();
    let mut seen = HashSet::new();

    for outcome in outcomes {
        diagnostics.merge(&outcome.diagnostics);
        let project = project_assignments
            .get(&outcome.report_session_id)
            .cloned()
            .unwrap_or_else(|| "미분류".to_string());
        for candidate in outcome.candidates {
            if !seen.insert(candidate.identity) {
                diagnostics.duplicate_events += 1;
                continue;
            }
            let rate = rate_for(&rates, &candidate.model, target_date);
            let event_totals = match Totals::from_usage(&candidate.usage, rate) {
                Ok(totals) => totals,
                Err(_) => {
                    diagnostics.invalid_token_events += 1;
                    continue;
                }
            };
            let label = model_label(&candidate.model).to_string();
            let project_sessions = sessions.entry(project.clone()).or_default();
            let session = project_sessions
                .entry(outcome.report_session_id.clone())
                .or_insert_with(|| SessionUsage {
                    session_id: outcome.report_session_id.clone(),
                    title: session_titles
                        .get(&outcome.report_session_id)
                        .cloned()
                        .unwrap_or_else(|| "제목 미확인".to_string()),
                    models: HashSet::new(),
                    totals: Totals::default(),
                });

            projects
                .entry(project.clone())
                .or_default()
                .merge(&event_totals);
            models
                .entry(label.clone())
                .or_default()
                .merge(&event_totals);
            session.models.insert(label);
            session.totals.merge(&event_totals);
            total.merge(&event_totals);
            diagnostics.aggregated_events += 1;
        }
    }

    Ok(Report {
        target_date,
        timezone_name,
        range_start,
        range_end,
        generated_at: Utc::now().with_timezone(&timezone),
        computer_name,
        projects,
        models,
        sessions,
        total,
        diagnostics,
    })
}

fn sum_totals<'a>(items: impl IntoIterator<Item = &'a Totals>) -> Totals {
    let mut total = Totals::default();
    for item in items {
        total.merge(item);
    }
    total
}

fn assert_same_totals(name: &str, candidate: &Totals, expected: &Totals) -> Result<()> {
    if candidate.total != expected.total {
        bail!("{name} total token mismatch");
    }
    if candidate.cached_input != expected.cached_input {
        bail!("{name} cached input mismatch");
    }
    if candidate.input != expected.input {
        bail!("{name} input mismatch");
    }
    if candidate.output != expected.output {
        bail!("{name} output mismatch");
    }
    if candidate.events != expected.events {
        bail!("{name} event count mismatch");
    }
    if (candidate.calculated_cost - expected.calculated_cost).abs() > 1e-9 {
        bail!("{name} calculated cost mismatch");
    }
    Ok(())
}

fn assert_report_integrity(report: &Report) -> Result<()> {
    let project_totals = sum_totals(report.projects.values());
    let model_totals = sum_totals(report.models.values());
    let session_totals = sum_totals(
        report
            .sessions
            .values()
            .flat_map(|sessions| sessions.values().map(|session| &session.totals)),
    );
    assert_same_totals("project", &project_totals, &report.total)?;
    assert_same_totals("model", &model_totals, &report.total)?;
    assert_same_totals("session", &session_totals, &report.total)?;

    for (project, project_total) in &report.projects {
        let candidate = sum_totals(
            report
                .sessions
                .get(project)
                .into_iter()
                .flat_map(|sessions| sessions.values().map(|session| &session.totals)),
        );
        assert_same_totals(&format!("{project} session"), &candidate, project_total)?;
    }
    Ok(())
}

fn sorted_totals(values: &HashMap<String, Totals>) -> Vec<(&String, &Totals)> {
    let mut rows: Vec<_> = values.iter().collect();
    rows.sort_by(|(left_name, left), (right_name, right)| {
        right
            .total
            .cmp(&left.total)
            .then_with(|| left_name.cmp(right_name))
    });
    rows
}

fn comma_integer(value: u64) -> String {
    let raw = value.to_string();
    let mut result = String::with_capacity(raw.len() + raw.len() / 3);
    for (index, character) in raw.chars().enumerate() {
        if index > 0 && (raw.len() - index).is_multiple_of(3) {
            result.push(',');
        }
        result.push(character);
    }
    result
}

fn comma_decimal(value: f64) -> String {
    let raw = format!("{value:.2}");
    let (integer, fraction) = raw.split_once('.').unwrap_or((&raw, "00"));
    let integer_value = integer.parse::<u64>().unwrap_or_default();
    format!("{}.{fraction}", comma_integer(integer_value))
}

fn format_tokens(value: u64) -> String {
    format!("{}M", comma_decimal(value as f64 / 1_000_000.0))
}

fn percent(value: u64, total: u64) -> String {
    if total == 0 {
        "0.0%".to_string()
    } else {
        format!("{:.1}%", value as f64 / total as f64 * 100.0)
    }
}

fn token_cell(value: u64, total: u64) -> String {
    format!("{} ({})", format_tokens(value), percent(value, total))
}

fn escape_markdown(value: &str) -> String {
    value.replace('|', "\\|").replace('\n', " ")
}

fn display_models(models: &HashSet<String>) -> String {
    MODEL_LABEL_ORDER
        .iter()
        .filter(|label| models.contains(**label))
        .copied()
        .collect::<Vec<_>>()
        .join(", ")
}

fn markdown_project_sessions(report: &Report) -> Vec<String> {
    let mut lines = vec![
        "| 프로젝트 / 세션 | 모델 | 총 토큰 | 캐시 입력 | 입력 | 출력 | 비용 |".to_string(),
        "|---|---|---:|---:|---:|---:|---:|".to_string(),
    ];
    for (project, project_total) in sorted_totals(&report.projects) {
        let empty_sessions = HashMap::new();
        let sessions = report.sessions.get(project).unwrap_or(&empty_sessions);
        lines.push(format!(
            "| **{} ({}개)** |  | **{}** | **{}** | **{}** | **{}** | **${:.2}** |",
            escape_markdown(project),
            sessions.len(),
            format_tokens(project_total.total),
            token_cell(project_total.cached_input, project_total.total),
            token_cell(project_total.input, project_total.total),
            token_cell(project_total.output, project_total.total),
            project_total.calculated_cost,
        ));
        let mut session_rows: Vec<_> = sessions.values().collect();
        session_rows.sort_by(|left, right| {
            right
                .totals
                .total
                .cmp(&left.totals.total)
                .then_with(|| left.title.cmp(&right.title))
                .then_with(|| left.session_id.cmp(&right.session_id))
        });
        for session in session_rows {
            let item = &session.totals;
            lines.push(format!(
                "| └ {} | {} | {} | {} | {} | {} | ${:.2} |",
                escape_markdown(&session.title),
                display_models(&session.models),
                format_tokens(item.total),
                token_cell(item.cached_input, item.total),
                token_cell(item.input, item.total),
                token_cell(item.output, item.total),
                item.calculated_cost,
            ));
        }
    }
    let session_count: usize = report.sessions.values().map(HashMap::len).sum();
    lines.push(format!(
        "| **전체 ({}개 세션)** |  | **{}** | **{}** | **{}** | **{}** | **${:.2}** |",
        session_count,
        format_tokens(report.total.total),
        token_cell(report.total.cached_input, report.total.total),
        token_cell(report.total.input, report.total.total),
        token_cell(report.total.output, report.total.total),
        report.total.calculated_cost,
    ));
    lines
}

fn markdown_table(values: &HashMap<String, Totals>, total: &Totals) -> Vec<String> {
    let mut lines = vec![
        "| 이름 | 총 토큰 | 캐시 입력 | 입력 | 출력 | 비용 |".to_string(),
        "|---|---:|---:|---:|---:|---:|".to_string(),
    ];
    for (name, item) in sorted_totals(values) {
        lines.push(format!(
            "| {} | {} | {} | {} | {} | ${:.2} |",
            name,
            format_tokens(item.total),
            token_cell(item.cached_input, item.total),
            token_cell(item.input, item.total),
            token_cell(item.output, item.total),
            item.calculated_cost,
        ));
    }
    lines.push(format!(
        "| 합계 | {} | {} | {} | {} | ${:.2} |",
        format_tokens(total.total),
        token_cell(total.cached_input, total.total),
        token_cell(total.input, total.total),
        token_cell(total.output, total.total),
        total.calculated_cost,
    ));
    lines
}

fn render_markdown(report: &Report) -> String {
    let diagnostics = &report.diagnostics;
    let session_count: usize = report.sessions.values().map(HashMap::len).sum();
    let mut lines = vec![
        format!("## {} Codex 일일 토큰 보고", report.target_date),
        String::new(),
        format!(
            "- 집계 시각: {} {}",
            report.generated_at.format("%Y-%m-%d %H:%M:%S"),
            report.timezone_name
        ),
        format!(
            "- 집계 기간: {} 이상, {} 미만 {}",
            report.range_start.format("%Y-%m-%d %H:%M:%S"),
            report.range_end.format("%Y-%m-%d %H:%M:%S"),
            report.timezone_name
        ),
        format!("- 집계 장치: {}", report.computer_name),
        format!(
            "- 원본 token_count 이벤트: {}개",
            comma_integer(diagnostics.original_events)
        ),
        format!(
            "- 중복 제거 이벤트: {}개",
            comma_integer(diagnostics.duplicate_events)
        ),
        format!(
            "- 상속 history 제외 이벤트: {}개",
            comma_integer(diagnostics.replayed_events)
        ),
        format!(
            "- 집계 token_count 이벤트: {}개",
            comma_integer(diagnostics.aggregated_events)
        ),
        format!(
            "- 프로젝트: {}개",
            comma_integer(report.projects.len() as u64)
        ),
        format!("- 세션: {}개", comma_integer(session_count as u64)),
        format!("- 모델: {}개", comma_integer(report.models.len() as u64)),
        String::new(),
        "### 프로젝트별".to_string(),
        String::new(),
    ];
    lines.extend(markdown_project_sessions(report));
    lines.push(String::new());
    lines.push("### 모델별".to_string());
    lines.push(String::new());
    lines.extend(markdown_table(&report.models, &report.total));
    lines.join("\n")
}

#[derive(Serialize)]
struct NamedTotals<'a> {
    name: &'a str,
    #[serde(flatten)]
    totals: &'a Totals,
}

#[derive(Serialize)]
struct SessionJson<'a> {
    session_id: &'a str,
    title: &'a str,
    models: Vec<&'static str>,
    #[serde(flatten)]
    totals: &'a Totals,
}

#[derive(Serialize)]
struct ProjectSessionsJson<'a> {
    project: &'a str,
    total: &'a Totals,
    sessions: Vec<SessionJson<'a>>,
}

#[derive(Serialize)]
struct RangeJson {
    start: String,
    end: String,
}

#[derive(Serialize)]
struct ReportJson<'a> {
    date: String,
    timezone: &'a str,
    range: RangeJson,
    generated_at: String,
    computer_name: &'a str,
    diagnostics: &'a Diagnostics,
    projects: Vec<NamedTotals<'a>>,
    models: Vec<NamedTotals<'a>>,
    sessions: Vec<ProjectSessionsJson<'a>>,
    total: &'a Totals,
}

fn report_to_json(report: &Report) -> Result<String> {
    let projects = sorted_totals(&report.projects)
        .into_iter()
        .map(|(name, totals)| NamedTotals { name, totals })
        .collect();
    let models = sorted_totals(&report.models)
        .into_iter()
        .map(|(name, totals)| NamedTotals { name, totals })
        .collect();
    let mut session_projects = Vec::new();
    for (project, project_total) in sorted_totals(&report.projects) {
        let mut sessions: Vec<_> = report
            .sessions
            .get(project)
            .into_iter()
            .flat_map(|items| items.values())
            .collect();
        sessions.sort_by(|left, right| {
            right
                .totals
                .total
                .cmp(&left.totals.total)
                .then_with(|| left.title.cmp(&right.title))
                .then_with(|| left.session_id.cmp(&right.session_id))
        });
        session_projects.push(ProjectSessionsJson {
            project,
            total: project_total,
            sessions: sessions
                .into_iter()
                .map(|session| SessionJson {
                    session_id: &session.session_id,
                    title: &session.title,
                    models: MODEL_LABEL_ORDER
                        .iter()
                        .filter(|label| session.models.contains(**label))
                        .copied()
                        .collect(),
                    totals: &session.totals,
                })
                .collect(),
        });
    }

    let payload = ReportJson {
        date: report.target_date.to_string(),
        timezone: &report.timezone_name,
        range: RangeJson {
            start: report.range_start.to_rfc3339(),
            end: report.range_end.to_rfc3339(),
        },
        generated_at: report.generated_at.to_rfc3339(),
        computer_name: &report.computer_name,
        diagnostics: &report.diagnostics,
        projects,
        models,
        sessions: session_projects,
        total: &report.total,
    };
    serde_json::to_string_pretty(&payload).context("JSON 보고서 직렬화 실패")
}

fn detect_computer_name() -> String {
    for (program, arguments) in [
        ("scutil", vec!["--get", "ComputerName"]),
        ("hostname", vec![]),
    ] {
        if let Ok(output) = Command::new(program).args(arguments).output()
            && output.status.success()
        {
            let value = String::from_utf8_lossy(&output.stdout).trim().to_string();
            if !value.is_empty() {
                return value;
            }
        }
    }
    "확인 불가".to_string()
}

fn default_rate_card() -> PathBuf {
    PathBuf::from(env!("CARGO_MANIFEST_DIR")).join("../../references/rate-card.toml")
}

fn home_path(path: &str) -> PathBuf {
    env::var_os("HOME")
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("."))
        .join(path)
}

fn parse_target_date(value: &str, timezone: Tz) -> Result<NaiveDate> {
    let today = Utc::now().with_timezone(&timezone).date_naive();
    match value {
        "today" => Ok(today),
        "yesterday" => today.pred_opt().context("전날 날짜 범위 초과"),
        explicit => NaiveDate::parse_from_str(explicit, "%Y-%m-%d")
            .context("--date는 today, yesterday 또는 YYYY-MM-DD여야 함"),
    }
}

fn run() -> Result<i32> {
    let cli = Cli::parse();
    let timezone = cli
        .timezone
        .parse::<Tz>()
        .with_context(|| format!("알 수 없는 timezone: {}", cli.timezone))?;
    let target_date = parse_target_date(&cli.date, timezone)?;
    let rate_card = cli.rate_card.unwrap_or_else(default_rate_card);
    let session_index = cli
        .session_index
        .unwrap_or_else(|| home_path(".codex/session_index.jsonl"));
    let global_state = cli
        .global_state
        .unwrap_or_else(|| home_path(".codex/.codex-global-state.json"));
    let roots = if let Some(root) = cli.sessions_root {
        vec![root]
    } else {
        let mut roots = vec![home_path(".codex/sessions")];
        let archived = home_path(".codex/archived_sessions");
        if archived.is_dir() {
            roots.push(archived);
        }
        roots
    };

    if !roots.first().is_some_and(|root| root.is_dir()) {
        bail!(
            "sessions root가 존재하지 않음: {}",
            roots.first().map_or_else(
                || "확인 불가".to_string(),
                |root| root.display().to_string()
            )
        );
    }
    if !rate_card.is_file() {
        bail!("rate card가 존재하지 않음: {}", rate_card.display());
    }

    let report = aggregate(
        &roots,
        target_date,
        timezone,
        cli.timezone,
        &rate_card,
        &session_index,
        &global_state,
        cli.computer_name.unwrap_or_else(detect_computer_name),
    )?;

    let mut failures = BTreeMap::new();
    if report.diagnostics.malformed_lines > 0 {
        failures.insert("malformed_lines", report.diagnostics.malformed_lines);
    }
    if report.diagnostics.unreadable_files > 0 {
        failures.insert("unreadable_files", report.diagnostics.unreadable_files);
    }
    if report.diagnostics.invalid_token_events > 0 {
        failures.insert(
            "invalid_token_events",
            report.diagnostics.invalid_token_events,
        );
    }
    if !failures.is_empty() {
        eprintln!(
            "집계 실패: 로그 또는 token_count metadata가 불완전함: {}",
            serde_json::to_string(&failures).unwrap_or_default()
        );
        return Ok(2);
    }

    assert_report_integrity(&report)?;
    match cli.format {
        OutputFormat::Markdown => println!("{}", render_markdown(&report)),
        OutputFormat::Json => println!("{}", report_to_json(&report)?),
    }
    Ok(0)
}

fn main() -> ExitCode {
    match run() {
        Ok(code) => ExitCode::from(code as u8),
        Err(error) => {
            eprintln!("집계 실패: {error:#}");
            ExitCode::from(2)
        }
    }
}
