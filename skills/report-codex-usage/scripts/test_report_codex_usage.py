from __future__ import annotations

import json
import tempfile
import unittest
from datetime import date
from pathlib import Path
from zoneinfo import ZoneInfo

from report_codex_usage import aggregate, assert_report_integrity

RATE_CARD = Path(__file__).resolve().parent.parent / "references" / "rate-card.toml"


def write_jsonl(path: Path, rows: list[dict]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        "".join(json.dumps(row, separators=(",", ":")) + "\n" for row in rows),
        encoding="utf-8",
    )


def token_event(
    timestamp: str,
    *,
    total: dict,
    last: dict,
) -> dict:
    return {
        "timestamp": timestamp,
        "type": "event_msg",
        "payload": {
            "type": "token_count",
            "info": {
                "total_token_usage": total,
                "last_token_usage": last,
            },
        },
    }


class AggregateTest(unittest.TestCase):
    def test_date_boundary_state_tracking_and_deduplication(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            agent_memory = root / "agent-memory"
            first_usage = {
                "input_tokens": 1_000,
                "cached_input_tokens": 700,
                "cache_write_input_tokens": 100,
                "output_tokens": 50,
            }
            second_usage = {
                "input_tokens": 500,
                "cached_input_tokens": 300,
                "cache_write_input_tokens": 50,
                "output_tokens": 25,
            }
            first_total = {**first_usage, "total_tokens": 1_050}
            second_total = {
                "input_tokens": 1_500,
                "cached_input_tokens": 1_000,
                "cache_write_input_tokens": 150,
                "output_tokens": 75,
                "total_tokens": 1_575,
            }

            first_rows = [
                {
                    "timestamp": "2026-07-30T14:59:59Z",
                    "type": "session_meta",
                    "payload": {
                        "id": "rollout-a",
                        "session_id": "thread-a",
                        "cwd": str(agent_memory),
                    },
                },
                token_event(
                    "2026-07-30T14:59:59Z",
                    total=first_total,
                    last=first_usage,
                ),
                token_event(
                    "2026-07-30T15:00:00Z",
                    total=first_total,
                    last=first_usage,
                ),
                {
                    "timestamp": "2026-07-30T16:00:00Z",
                    "type": "event_msg",
                    "payload": {
                        "type": "thread_settings_applied",
                        "thread_settings": {
                            "cwd": str(root / "project-one"),
                            "model": "gpt-5.6-luna",
                        },
                    },
                },
                token_event(
                    "2026-07-30T16:00:01Z",
                    total=second_total,
                    last=second_usage,
                ),
                {
                    "timestamp": "2026-07-31T14:59:59Z",
                    "type": "turn_context",
                    "payload": {
                        "cwd": str(root / "project-one"),
                        "model": "gpt-5.6-luna",
                    },
                },
                token_event(
                    "2026-07-31T15:00:00Z",
                    total=second_total,
                    last=second_usage,
                ),
            ]
            duplicate_rows = [
                {
                    "timestamp": "2026-07-30T14:59:59Z",
                    "type": "session_meta",
                    "payload": {
                        "id": "rollout-b",
                        "session_id": "thread-a",
                        "cwd": str(agent_memory),
                    },
                },
                {
                    "timestamp": "2026-07-30T15:00:00Z",
                    "type": "turn_context",
                    "payload": {
                        "cwd": str(agent_memory),
                        "model": "gpt-5.6-luna",
                    },
                },
                token_event(
                    "2026-07-30T15:00:00Z",
                    total=first_total,
                    last=first_usage,
                ),
            ]
            write_jsonl(root / "a.jsonl", first_rows)
            write_jsonl(root / "b.jsonl", duplicate_rows)

            report = aggregate(
                sessions_root=root,
                target_date=date(2026, 7, 31),
                timezone_info=ZoneInfo("Asia/Seoul"),
                timezone_name="Asia/Seoul",
                rate_card=RATE_CARD,
                agent_memory_root=agent_memory,
                computer_name="Fixture Mac",
            )

            assert_report_integrity(report)
            self.assertEqual(report.diagnostics.original_events, 3)
            self.assertEqual(report.diagnostics.duplicate_events, 1)
            self.assertEqual(report.diagnostics.aggregated_events, 2)
            self.assertEqual(report.total.total, 1_575)
            self.assertEqual(report.total.cached_input, 1_000)
            self.assertEqual(report.total.input, 500)
            self.assertEqual(report.total.output, 75)
            self.assertEqual(set(report.projects), {"agent-memory", "project-one"})
            self.assertEqual(set(report.models), {"gpt-5.6-luna"})

            expected_cost = (
                (200 * 0.20 + 700 * 0.02 + 100 * 0.25 + 50 * 1.20)
                + (150 * 0.20 + 300 * 0.02 + 50 * 0.25 + 25 * 1.20)
            ) / 1_000_000
            self.assertAlmostEqual(report.total.calculated_cost, expected_cost)


if __name__ == "__main__":
    unittest.main()
