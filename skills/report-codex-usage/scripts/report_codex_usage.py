from __future__ import annotations

import argparse
import json
import platform
import subprocess
import sys
from collections import defaultdict
from dataclasses import asdict, dataclass, field
from datetime import date, datetime, time, timedelta, timezone
from pathlib import Path
from typing import Any
from zoneinfo import ZoneInfo, ZoneInfoNotFoundError

import tomllib

SKILL_ROOT = Path(__file__).resolve().parent.parent
DEFAULT_RATE_CARD = SKILL_ROOT / "references" / "rate-card.toml"
DEFAULT_SESSIONS_ROOT = Path.home() / ".codex" / "sessions"
DEFAULT_ARCHIVED_SESSIONS_ROOT = Path.home() / ".codex" / "archived_sessions"
DEFAULT_SESSION_INDEX = Path.home() / ".codex" / "session_index.jsonl"
DEFAULT_GLOBAL_STATE = Path.home() / ".codex" / ".codex-global-state.json"
MODEL_LABEL_ORDER = ("sol", "terra", "luna", "review", "other")


@dataclass(frozen=True)
class Rate:
    model: str
    effective_from: date
    input_per_million: float
    cached_input_per_million: float
    output_per_million: float
    cache_write_per_million: float


@dataclass
class Totals:
    total: int = 0
    cached_input: int = 0
    input: int = 0
    output: int = 0
    calculated_cost: float = 0.0
    events: int = 0

    def add(self, usage: dict[str, Any], rate: Rate | None) -> None:
        input_tokens = integer_token(usage, "input_tokens")
        cached_input = integer_token(usage, "cached_input_tokens")
        cache_write = integer_token(usage, "cache_write_input_tokens")
        output = integer_token(usage, "output_tokens")

        if cached_input + cache_write > input_tokens:
            raise ValueError(
                "cached_input_tokens + cache_write_input_tokens exceeds input_tokens"
            )

        self.total += input_tokens + output
        self.cached_input += cached_input
        self.input += input_tokens - cached_input
        self.output += output
        self.events += 1

        if rate is None:
            return

        regular_input = input_tokens - cached_input - cache_write
        self.calculated_cost += (
            regular_input * rate.input_per_million
            + cached_input * rate.cached_input_per_million
            + cache_write * rate.cache_write_per_million
            + output * rate.output_per_million
        ) / 1_000_000

    def merge(self, other: Totals) -> None:
        self.total += other.total
        self.cached_input += other.cached_input
        self.input += other.input
        self.output += other.output
        self.calculated_cost += other.calculated_cost
        self.events += other.events


@dataclass
class Diagnostics:
    files_scanned: int = 0
    files_with_target_events: int = 0
    original_events: int = 0
    duplicate_events: int = 0
    replayed_events: int = 0
    aggregated_events: int = 0
    malformed_lines: int = 0
    unreadable_files: int = 0
    token_events_without_usage: int = 0
    invalid_token_events: int = 0


@dataclass
class SessionUsage:
    session_id: str
    title: str
    models: set[str] = field(default_factory=set)
    totals: Totals = field(default_factory=Totals)


@dataclass
class Report:
    target_date: date
    timezone_name: str
    range_start: datetime
    range_end: datetime
    generated_at: datetime
    computer_name: str
    projects: dict[str, Totals]
    models: dict[str, Totals]
    sessions: dict[str, dict[str, SessionUsage]]
    total: Totals
    diagnostics: Diagnostics


def integer_token(usage: dict[str, Any], key: str) -> int:
    value = usage.get(key, 0)
    if value is None:
        return 0
    if isinstance(value, bool) or not isinstance(value, int) or value < 0:
        raise ValueError(f"{key} must be a non-negative integer")
    return value


def parse_target_date(value: str, timezone_info: ZoneInfo) -> date:
    today = datetime.now(timezone_info).date()
    if value == "today":
        return today
    if value == "yesterday":
        return today - timedelta(days=1)
    try:
        return date.fromisoformat(value)
    except ValueError as error:
        raise argparse.ArgumentTypeError(
            "--date must be today, yesterday, or YYYY-MM-DD"
        ) from error


def parse_timestamp(value: Any) -> datetime | None:
    if not isinstance(value, str):
        return None
    normalized = value[:-1] + "+00:00" if value.endswith("Z") else value
    try:
        parsed = datetime.fromisoformat(normalized)
    except ValueError:
        return None
    if parsed.tzinfo is None:
        return parsed.replace(tzinfo=timezone.utc)
    return parsed.astimezone(timezone.utc)


def stable_json(value: Any) -> str:
    return json.dumps(
        value,
        ensure_ascii=False,
        sort_keys=True,
        separators=(",", ":"),
    )


def load_rates(path: Path) -> dict[str, list[Rate]]:
    with path.open("rb") as handle:
        data = tomllib.load(handle)

    rates: dict[str, list[Rate]] = defaultdict(list)
    for raw in data.get("rate", []):
        rate = Rate(
            model=raw["model"],
            effective_from=date.fromisoformat(raw["effective_from"]),
            input_per_million=float(raw["input_per_million"]),
            cached_input_per_million=float(raw["cached_input_per_million"]),
            output_per_million=float(raw["output_per_million"]),
            cache_write_per_million=float(raw["cache_write_per_million"]),
        )
        rates[rate.model].append(rate)

    for model_rates in rates.values():
        model_rates.sort(key=lambda item: item.effective_from)
    return dict(rates)


def rate_for(
    rates: dict[str, list[Rate]],
    model: str,
    target_date: date,
) -> Rate | None:
    candidates = [
        rate for rate in rates.get(model, []) if rate.effective_from <= target_date
    ]
    return candidates[-1] if candidates else None


def load_session_titles(path: Path) -> dict[str, str]:
    titles: dict[str, str] = {}
    try:
        handle = path.open("r", encoding="utf-8")
    except OSError:
        return titles

    with handle:
        for line in handle:
            try:
                row = json.loads(line)
            except (json.JSONDecodeError, UnicodeDecodeError):
                continue
            session_id = row.get("id")
            title = row.get("thread_name")
            if isinstance(session_id, str) and isinstance(title, str) and title:
                titles[session_id] = title
    return titles


def load_project_assignments(path: Path) -> dict[str, str]:
    try:
        with path.open("r", encoding="utf-8") as handle:
            state = json.load(handle)
    except (OSError, json.JSONDecodeError, UnicodeDecodeError):
        return {}

    raw_projects = state.get("local-projects")
    raw_assignments = state.get("thread-project-assignments")
    if not isinstance(raw_projects, dict) or not isinstance(raw_assignments, dict):
        return {}

    project_names: dict[str, str] = {}
    for project_id, project in raw_projects.items():
        if not isinstance(project_id, str) or not isinstance(project, dict):
            continue
        name = project.get("name")
        if isinstance(name, str) and name:
            project_names[project_id] = name

    assignments: dict[str, str] = {}
    for thread_id, assignment in raw_assignments.items():
        if not isinstance(thread_id, str) or not isinstance(assignment, dict):
            continue
        if assignment.get("projectKind") != "local":
            continue
        project_id = assignment.get("projectId")
        if isinstance(project_id, str) and project_id in project_names:
            assignments[thread_id] = project_names[project_id]
    return assignments


def model_label(model: str) -> str:
    normalized = model.lower()
    if normalized in {"gpt-5.6", "gpt-5.6-sol"} or "5.6-sol" in normalized:
        return "sol"
    if "terra" in normalized:
        return "terra"
    if "luna" in normalized:
        return "luna"
    if normalized == "codex-auto-review" or "auto-review" in normalized:
        return "review"
    return "other"


def display_models(models: set[str]) -> str:
    return ", ".join(label for label in MODEL_LABEL_ORDER if label in models)


def projected_records(
    path: Path,
    diagnostics: Diagnostics,
) -> tuple[list[tuple[str, datetime | None, dict[str, Any]]], set[str]]:
    records: list[tuple[str, datetime | None, dict[str, Any]]] = []
    models: set[str] = set()

    try:
        handle = path.open("r", encoding="utf-8")
    except OSError:
        diagnostics.unreadable_files += 1
        return records, models

    with handle:
        for line in handle:
            try:
                row = json.loads(line)
            except (json.JSONDecodeError, UnicodeDecodeError):
                diagnostics.malformed_lines += 1
                continue

            row_type = row.get("type")
            payload = row.get("payload")
            if not isinstance(payload, dict):
                payload = {}
            timestamp = parse_timestamp(row.get("timestamp"))

            if row_type == "session_meta":
                records.append(
                    (
                        "session_meta",
                        timestamp,
                        {
                            "id": payload.get("id"),
                            "session_id": payload.get("session_id"),
                        },
                    )
                )
                continue

            if row_type == "turn_context":
                model = payload.get("model")
                if isinstance(model, str) and model:
                    models.add(model)
                records.append(
                    (
                        "turn_context",
                        timestamp,
                        {"model": model},
                    )
                )
                continue

            if row_type != "event_msg":
                continue

            payload_type = payload.get("type")
            if payload_type == "thread_settings_applied":
                settings = payload.get("thread_settings")
                if not isinstance(settings, dict):
                    settings = {}
                model = settings.get("model")
                if isinstance(model, str) and model:
                    models.add(model)
                records.append(
                    (
                        "thread_settings",
                        timestamp,
                        {"model": model},
                    )
                )
            elif payload_type == "task_started":
                records.append(("task_started", timestamp, {}))
            elif payload_type == "token_count":
                records.append(
                    (
                        "token_count",
                        timestamp,
                        {"info": payload.get("info")},
                    )
                )

    return records, models


def aggregate(
    *,
    sessions_root: Path,
    additional_sessions_roots: tuple[Path, ...] = (),
    target_date: date,
    timezone_info: ZoneInfo,
    timezone_name: str,
    rate_card: Path,
    session_index: Path,
    global_state: Path,
    computer_name: str,
) -> Report:
    range_start = datetime.combine(target_date, time.min, timezone_info)
    range_end = range_start + timedelta(days=1)
    range_start_utc = range_start.astimezone(timezone.utc)
    range_end_utc = range_end.astimezone(timezone.utc)

    rates = load_rates(rate_card)
    session_titles = load_session_titles(session_index)
    project_assignments = load_project_assignments(global_state)
    projects: dict[str, Totals] = defaultdict(Totals)
    models: dict[str, Totals] = defaultdict(Totals)
    sessions: dict[str, dict[str, SessionUsage]] = defaultdict(dict)
    overall = Totals()
    diagnostics = Diagnostics()
    seen: set[tuple[str, str, str, str]] = set()

    files = sorted(
        path
        for root in (sessions_root, *additional_sessions_roots)
        for path in root.rglob("*.jsonl")
    )
    diagnostics.files_scanned = len(files)

    for path in files:
        records, file_models = projected_records(path, diagnostics)
        unique_file_model = next(iter(file_models)) if len(file_models) == 1 else None
        first_meta = next(
            (data for kind, _, data in records if kind == "session_meta"),
            {},
        )
        rollout_id = first_meta.get("id")
        if not isinstance(rollout_id, str) or not rollout_id:
            rollout_id = path.stem
        report_session_id = first_meta.get("session_id") or rollout_id
        if not isinstance(report_session_id, str) or not report_session_id:
            report_session_id = rollout_id

        replay_cutoff: int | None = None
        replay_only_rollout = False
        last_foreign_meta = max(
            (
                index
                for index, (kind, _, data) in enumerate(records)
                if kind == "session_meta" and data.get("id") != rollout_id
            ),
            default=None,
        )
        if last_foreign_meta is not None:
            replay_cutoff = next(
                (
                    index
                    for index, (kind, _, _) in enumerate(records)
                    if index > last_foreign_meta and kind == "task_started"
                ),
                None,
            )
            replay_only_rollout = replay_cutoff is None

        current_session_id = report_session_id
        current_model: str | None = None
        file_has_target_event = False

        for record_index, (kind, timestamp, data) in enumerate(records):
            if kind == "session_meta":
                continue

            if kind in {"turn_context", "thread_settings"}:
                model = data.get("model")
                if isinstance(model, str) and model:
                    current_model = model
                continue

            if kind != "token_count":
                continue

            if (
                timestamp is None
                or timestamp < range_start_utc
                or timestamp >= range_end_utc
            ):
                continue

            diagnostics.original_events += 1
            file_has_target_event = True
            if replay_only_rollout or (
                replay_cutoff is not None and record_index < replay_cutoff
            ):
                diagnostics.replayed_events += 1
                continue
            info = data.get("info")
            if not isinstance(info, dict):
                diagnostics.token_events_without_usage += 1
                continue

            last_usage = info.get("last_token_usage")
            if not isinstance(last_usage, dict):
                diagnostics.token_events_without_usage += 1
                continue

            session_id = current_session_id
            total_usage = info.get("total_token_usage")
            identity = (
                rollout_id,
                timestamp.isoformat(),
                stable_json(total_usage),
                stable_json(last_usage),
            )
            if identity in seen:
                diagnostics.duplicate_events += 1
                continue
            seen.add(identity)

            event_model = current_model or unique_file_model or "미분류"
            event_model_label = model_label(event_model)
            event_project = project_assignments.get(report_session_id, "미분류")
            event_rate = rate_for(rates, event_model, target_date)
            event_totals = Totals()
            try:
                event_totals.add(last_usage, event_rate)
            except ValueError:
                diagnostics.invalid_token_events += 1
                continue

            session = sessions[event_project].setdefault(
                session_id,
                SessionUsage(
                    session_id=session_id,
                    title=session_titles.get(session_id, "제목 미확인"),
                ),
            )
            projects[event_project].merge(event_totals)
            models[event_model_label].merge(event_totals)
            session.totals.merge(event_totals)
            overall.merge(event_totals)
            session.models.add(event_model_label)
            diagnostics.aggregated_events += 1

        if file_has_target_event:
            diagnostics.files_with_target_events += 1

    return Report(
        target_date=target_date,
        timezone_name=timezone_name,
        range_start=range_start,
        range_end=range_end,
        generated_at=datetime.now(timezone_info),
        computer_name=computer_name,
        projects=dict(projects),
        models=dict(models),
        sessions={project: dict(items) for project, items in sessions.items()},
        total=overall,
        diagnostics=diagnostics,
    )


def assert_report_integrity(report: Report) -> None:
    project_totals = sum_totals(report.projects.values())
    model_totals = sum_totals(report.models.values())
    session_totals = sum_totals(
        session.totals
        for sessions in report.sessions.values()
        for session in sessions.values()
    )
    for name, candidate in (
        ("project", project_totals),
        ("model", model_totals),
        ("session", session_totals),
    ):
        if candidate.total != report.total.total:
            raise RuntimeError(f"{name} total token mismatch")
        if candidate.cached_input != report.total.cached_input:
            raise RuntimeError(f"{name} cached input mismatch")
        if candidate.input != report.total.input:
            raise RuntimeError(f"{name} input mismatch")
        if candidate.output != report.total.output:
            raise RuntimeError(f"{name} output mismatch")
        if abs(candidate.calculated_cost - report.total.calculated_cost) > 1e-9:
            raise RuntimeError(f"{name} calculated cost mismatch")

    for project, project_total in report.projects.items():
        candidate = sum_totals(
            session.totals for session in report.sessions.get(project, {}).values()
        )
        for field_name in ("total", "cached_input", "input", "output", "events"):
            if getattr(candidate, field_name) != getattr(project_total, field_name):
                raise RuntimeError(f"{project} session {field_name} mismatch")
        if abs(candidate.calculated_cost - project_total.calculated_cost) > 1e-9:
            raise RuntimeError(f"{project} session calculated cost mismatch")


def sum_totals(items: Any) -> Totals:
    result = Totals()
    for item in items:
        result.merge(item)
    return result


def sorted_totals(values: dict[str, Totals]) -> list[tuple[str, Totals]]:
    return sorted(values.items(), key=lambda item: (-item[1].total, item[0]))


def format_tokens(value: int) -> str:
    return f"{value / 1_000_000:,.2f}M"


def percent(value: int, total: int) -> str:
    if total == 0:
        return "0.0%"
    return f"{value / total * 100:.1f}%"


def token_cell(value: int, total: int) -> str:
    return f"{format_tokens(value)} ({percent(value, total)})"


def escape_markdown(value: str) -> str:
    return value.replace("|", "\\|").replace("\n", " ")


def markdown_table(values: dict[str, Totals], total: Totals) -> list[str]:
    lines = [
        "| 이름 | 총 토큰 | 캐시 입력 | 입력 | 출력 | 비용 |",
        "|---|---:|---:|---:|---:|---:|",
    ]
    for name, item in sorted_totals(values):
        lines.append(
            "| "
            + " | ".join(
                [
                    name,
                    format_tokens(item.total),
                    token_cell(item.cached_input, item.total),
                    token_cell(item.input, item.total),
                    token_cell(item.output, item.total),
                    f"${item.calculated_cost:.2f}",
                ]
            )
            + " |"
        )
    lines.append(
        "| "
        + " | ".join(
            [
                "합계",
                format_tokens(total.total),
                token_cell(total.cached_input, total.total),
                token_cell(total.input, total.total),
                token_cell(total.output, total.total),
                f"${total.calculated_cost:.2f}",
            ]
        )
        + " |"
    )
    return lines


def markdown_project_sessions(report: Report) -> list[str]:
    lines = [
        "| 프로젝트 / 세션 | 모델 | 총 토큰 | 캐시 입력 | 입력 | 출력 | 비용 |",
        "|---|---|---:|---:|---:|---:|---:|",
    ]
    for project, project_total in sorted_totals(report.projects):
        sessions = report.sessions.get(project, {})
        lines.append(
            "| "
            + " | ".join(
                [
                    f"**{escape_markdown(project)} ({len(sessions)}개)**",
                    "",
                    f"**{format_tokens(project_total.total)}**",
                    f"**{token_cell(project_total.cached_input, project_total.total)}**",
                    f"**{token_cell(project_total.input, project_total.total)}**",
                    f"**{token_cell(project_total.output, project_total.total)}**",
                    f"**${project_total.calculated_cost:.2f}**",
                ]
            )
            + " |"
        )
        for session in sorted(
            sessions.values(),
            key=lambda item: (-item.totals.total, item.title, item.session_id),
        ):
            item = session.totals
            lines.append(
                "| "
                + " | ".join(
                    [
                        f"└ {escape_markdown(session.title)}",
                        display_models(session.models),
                        format_tokens(item.total),
                        token_cell(item.cached_input, item.total),
                        token_cell(item.input, item.total),
                        token_cell(item.output, item.total),
                        f"${item.calculated_cost:.2f}",
                    ]
                )
                + " |"
            )
    lines.append(
        "| "
        + " | ".join(
            [
                f"**전체 ({sum(len(items) for items in report.sessions.values())}개 세션)**",
                "",
                f"**{format_tokens(report.total.total)}**",
                f"**{token_cell(report.total.cached_input, report.total.total)}**",
                f"**{token_cell(report.total.input, report.total.total)}**",
                f"**{token_cell(report.total.output, report.total.total)}**",
                f"**${report.total.calculated_cost:.2f}**",
            ]
        )
        + " |"
    )
    return lines


def render_markdown(report: Report) -> str:
    diagnostics = report.diagnostics
    lines = [
        f"## {report.target_date.isoformat()} Codex 일일 토큰 보고",
        "",
        f"- 집계 시각: {report.generated_at:%Y-%m-%d %H:%M:%S} {report.timezone_name}",
        (
            f"- 집계 기간: {report.range_start:%Y-%m-%d %H:%M:%S} 이상, "
            f"{report.range_end:%Y-%m-%d %H:%M:%S} 미만 {report.timezone_name}"
        ),
        f"- 집계 장치: {report.computer_name}",
        f"- 원본 token_count 이벤트: {diagnostics.original_events:,}개",
        f"- 중복 제거 이벤트: {diagnostics.duplicate_events:,}개",
        f"- 상속 history 제외 이벤트: {diagnostics.replayed_events:,}개",
        f"- 집계 token_count 이벤트: {diagnostics.aggregated_events:,}개",
        f"- 프로젝트: {len(report.projects):,}개",
        f"- 세션: {sum(len(items) for items in report.sessions.values()):,}개",
        f"- 모델: {len(report.models):,}개",
        "",
        "### 프로젝트별",
        "",
        *markdown_project_sessions(report),
        "",
        "### 모델별",
        "",
        *markdown_table(report.models, report.total),
    ]
    return "\n".join(lines)


def report_to_json(report: Report) -> str:
    def serialize_totals(values: dict[str, Totals]) -> list[dict[str, Any]]:
        return [
            {"name": name, **asdict(total)} for name, total in sorted_totals(values)
        ]

    def serialize_sessions() -> list[dict[str, Any]]:
        projects: list[dict[str, Any]] = []
        for project, project_total in sorted_totals(report.projects):
            sessions = sorted(
                report.sessions.get(project, {}).values(),
                key=lambda item: (-item.totals.total, item.title, item.session_id),
            )
            projects.append(
                {
                    "project": project,
                    "total": asdict(project_total),
                    "sessions": [
                        {
                            "session_id": session.session_id,
                            "title": session.title,
                            "models": [
                                label
                                for label in MODEL_LABEL_ORDER
                                if label in session.models
                            ],
                            **asdict(session.totals),
                        }
                        for session in sessions
                    ],
                }
            )
        return projects

    payload = {
        "date": report.target_date.isoformat(),
        "timezone": report.timezone_name,
        "range": {
            "start": report.range_start.isoformat(),
            "end": report.range_end.isoformat(),
        },
        "generated_at": report.generated_at.isoformat(),
        "computer_name": report.computer_name,
        "diagnostics": asdict(report.diagnostics),
        "projects": serialize_totals(report.projects),
        "models": serialize_totals(report.models),
        "sessions": serialize_sessions(),
        "total": asdict(report.total),
    }
    return json.dumps(payload, ensure_ascii=False, indent=2)


def detect_computer_name() -> str:
    try:
        result = subprocess.run(
            ["scutil", "--get", "ComputerName"],
            check=False,
            capture_output=True,
            text=True,
        )
        if result.stdout.strip():
            return result.stdout.strip()
    except OSError:
        pass
    return platform.node() or "확인 불가"


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(
        description="Aggregate daily Codex token usage from local session JSONL metadata."
    )
    parser.add_argument(
        "--date",
        default="yesterday",
        help="today, yesterday, or YYYY-MM-DD (default: yesterday)",
    )
    parser.add_argument(
        "--timezone",
        default="Asia/Seoul",
        help="IANA timezone used for the calendar-day boundary",
    )
    parser.add_argument(
        "--sessions-root",
        type=Path,
        help="override the default active and archived Codex session directories",
    )
    parser.add_argument(
        "--rate-card",
        type=Path,
        default=DEFAULT_RATE_CARD,
        help="TOML rate card",
    )
    parser.add_argument(
        "--session-index",
        type=Path,
        default=DEFAULT_SESSION_INDEX,
        help="Codex session title index",
    )
    parser.add_argument(
        "--global-state",
        type=Path,
        default=DEFAULT_GLOBAL_STATE,
        help="Codex desktop global state containing UI project assignments",
    )
    parser.add_argument(
        "--computer-name",
        help="Override the detected computer name",
    )
    parser.add_argument(
        "--format",
        choices=("markdown", "json"),
        default="markdown",
    )
    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()

    try:
        timezone_info = ZoneInfo(args.timezone)
    except ZoneInfoNotFoundError:
        parser.error(f"unknown timezone: {args.timezone}")

    target_date = parse_target_date(args.date, timezone_info)
    if args.sessions_root is None:
        sessions_root = DEFAULT_SESSIONS_ROOT
        additional_sessions_roots = tuple(
            root for root in (DEFAULT_ARCHIVED_SESSIONS_ROOT,) if root.is_dir()
        )
    else:
        sessions_root = args.sessions_root.expanduser()
        additional_sessions_roots = ()
    rate_card = args.rate_card.expanduser()
    session_index = args.session_index.expanduser()
    global_state = args.global_state.expanduser()

    if not sessions_root.is_dir():
        parser.error(f"sessions root does not exist: {sessions_root}")
    if not rate_card.is_file():
        parser.error(f"rate card does not exist: {rate_card}")

    report = aggregate(
        sessions_root=sessions_root,
        additional_sessions_roots=additional_sessions_roots,
        target_date=target_date,
        timezone_info=timezone_info,
        timezone_name=args.timezone,
        rate_card=rate_card,
        session_index=session_index,
        global_state=global_state,
        computer_name=args.computer_name or detect_computer_name(),
    )

    diagnostics = report.diagnostics
    failures = {
        "malformed_lines": diagnostics.malformed_lines,
        "unreadable_files": diagnostics.unreadable_files,
        "invalid_token_events": diagnostics.invalid_token_events,
    }
    failures = {name: count for name, count in failures.items() if count}
    if failures:
        print(
            "집계 실패: 로그 또는 token_count metadata가 불완전함: "
            + stable_json(failures),
            file=sys.stderr,
        )
        return 2

    assert_report_integrity(report)
    if args.format == "json":
        print(report_to_json(report))
    else:
        print(render_markdown(report))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
