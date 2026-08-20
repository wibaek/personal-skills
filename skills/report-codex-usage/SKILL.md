---
name: report-codex-usage
description: Aggregate local Codex session JSONL token metadata into reproducible daily project, session, and model reports with Standard API equivalent costs. Use for Codex token usage reports, yesterday/today or historical date analysis, project/session/model breakdowns, duplicate-event checks, and scheduled Codex usage reporting without reading message, prompt, or response bodies.
---

# Report Codex Usage

Generate reports with the bundled deterministic script. Do not recreate the aggregation logic in the prompt.

## Run a report

Run from the skill directory:

```bash
scripts/report_codex_usage --date yesterday --timezone Asia/Seoul --format markdown
```

Use an explicit date for reproducible historical reports:

```bash
scripts/report_codex_usage --date 2026-07-31 --timezone Asia/Seoul --format markdown
```

The launcher builds the bundled Rust aggregator on its first run and reuses the cached release build afterward.

Use `--format json` when machine-readable totals or further analysis is required. Pass `--sessions-root` only when reading a non-default Codex session directory or a test fixture. Session titles come from `~/.codex/session_index.jsonl`; use `--session-index` only for a non-default index or fixture. Project assignments come from `~/.codex/.codex-global-state.json`; use `--global-state` only for a non-default state file or fixture.

Return the script output without recalculating token counts, percentages, costs, or event counts in the model response.

## Report format

- Show each project aggregate first, followed by its indented session rows.
- Resolve a task's project from the current Codex UI assignment in `thread-project-assignments`, then use the matching `local-projects` name. Attribute subagent usage to its root task's assignment. Use `미분류` when no retained assignment exists.
- Resolve session titles from the session index. Use `제목 미확인` when no title is available.
- Normalize model display labels to `sol`, `terra`, `luna`, `5.5`, `5.4`, `review`, or `other`.
- Join multiple model labels in the fixed order above with `, `.
- Show token values in millions and USD costs with two decimal places.
- Show cached input, input, and output percentages relative to each row's total in both project/session and model tables.

## Data handling

- Read only local `~/.codex/sessions/**/*.jsonl`, `~/.codex/archived_sessions/**/*.jsonl`, `~/.codex/session_index.jsonl`, and `~/.codex/.codex-global-state.json` unless the user provides another root.
- Access only the fields projected by the script from `session_meta`, `turn_context`, `event_msg.thread_settings_applied`, and `event_msg.token_count`.
- Access only `id` and `thread_name` from the session index. When a session has multiple title records, use the last title.
- Access only `local-projects[*].name` and each `thread-project-assignments` entry's `projectKind` and `projectId` from the global state. Treat it as the current assignment, including for historical usage; do not infer a project from the task cwd.
- Never print, summarize, retain, or reason about message, prompt, response, or tool-call bodies.
- Keep the report read-only. Do not create report files unless the user explicitly requests one.

## Validation and failures

- Treat a nonzero script exit as a failed report. Report the error and do not estimate missing values.
- Ignore rate-limit status events whose `token_count.info` is null because they contain no usage delta.
- Exclude inherited history `token_count` records replayed at the start of any rollout file that embeds another session's metadata. Count only token events produced after the current task actually starts.
- Preserve the half-open date range shown by the script.
- Confirm project, session, and model totals match the overall token and calculated-cost totals.
- Keep models without a configured rate in token totals. Do not substitute another model's price.
- Update [references/rate-card.toml](references/rate-card.toml) only from an explicitly approved pricing source.

## Scheduled automation

Keep the schedule prompt short and invoke this skill explicitly:

```text
$report-codex-usage 스킬을 사용해 Asia/Seoul 기준 전날 Codex 사용량을 프로젝트별·세션별·모델별로 집계하고 이 General task에 한국어로 보고한다. 파일은 수정하지 않는다.
```

The schedule controls when the report runs. The script controls date boundaries, deduplication, aggregation, pricing, formatting, and validation.
