---
name: report-codex-usage
description: Aggregate local Codex session JSONL token metadata into reproducible daily project and model reports with Standard API equivalent costs. Use for Codex token usage reports, yesterday/today or historical date analysis, project/model breakdowns, duplicate-event checks, and scheduled Codex usage reporting without reading message, prompt, or response bodies.
---

# Report Codex Usage

Generate reports with the bundled deterministic script. Do not recreate the aggregation logic in the prompt.

## Run a report

Run from the skill directory:

```bash
python3 scripts/report_codex_usage.py --date yesterday --timezone Asia/Seoul --format markdown
```

Use an explicit date for reproducible historical reports:

```bash
python3 scripts/report_codex_usage.py --date 2026-07-31 --timezone Asia/Seoul --format markdown
```

Use `--format json` when machine-readable totals or further analysis is required. Pass `--sessions-root` only when reading a non-default Codex session directory or a test fixture.

Return the script output without recalculating token counts, percentages, costs, or event counts in the model response.

## Data handling

- Read only local `~/.codex/sessions/**/*.jsonl` unless the user provides another root.
- Access only the fields projected by the script from `session_meta`, `turn_context`, `event_msg.thread_settings_applied`, and `event_msg.token_count`.
- Never print, summarize, retain, or reason about message, prompt, response, or tool-call bodies.
- Keep the report read-only. Do not create report files unless the user explicitly requests one.

## Validation and failures

- Treat a nonzero script exit as a failed report. Report the error and do not estimate missing values.
- Preserve the half-open date range shown by the script.
- Confirm project and model totals match the overall token and calculated-cost totals.
- Keep models without a configured rate in token totals. Do not substitute another model's price.
- Update [references/rate-card.toml](references/rate-card.toml) only from an explicitly approved pricing source.

## Scheduled automation

Keep the schedule prompt short and invoke this skill explicitly:

```text
$report-codex-usage 스킬을 사용해 Asia/Seoul 기준 전날 Codex 사용량을 프로젝트별·모델별로 집계하고 이 General task에 한국어로 보고한다. 파일은 수정하지 않는다.
```

The schedule controls when the report runs. The script controls date boundaries, deduplication, aggregation, pricing, formatting, and validation.
