from __future__ import annotations

import importlib.util
import json
import sys
import tempfile
import unittest
from datetime import date
from pathlib import Path
from zoneinfo import ZoneInfo

SCRIPT = Path(__file__).parents[1] / "scripts" / "report_codex_usage.py"
SPEC = importlib.util.spec_from_file_location("report_codex_usage", SCRIPT)
if SPEC is None or SPEC.loader is None:
    raise RuntimeError(f"cannot load {SCRIPT}")
usage_report = importlib.util.module_from_spec(SPEC)
sys.modules[SPEC.name] = usage_report
SPEC.loader.exec_module(usage_report)


def write_jsonl(path: Path, rows: list[dict[str, object]]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(
        "\n".join(json.dumps(row, ensure_ascii=False) for row in rows) + "\n",
        encoding="utf-8",
    )


def token_count(
    timestamp: str,
    *,
    input_tokens: int,
    cached_input_tokens: int,
    output_tokens: int,
) -> dict[str, object]:
    usage = {
        "input_tokens": input_tokens,
        "cached_input_tokens": cached_input_tokens,
        "output_tokens": output_tokens,
    }
    return {
        "timestamp": timestamp,
        "type": "event_msg",
        "payload": {
            "type": "token_count",
            "info": {
                "total_token_usage": usage,
                "last_token_usage": usage,
            },
        },
    }


class ModelDisplayTests(unittest.TestCase):
    def test_model_labels_use_five_display_groups(self) -> None:
        cases = {
            "gpt-5.6-sol": "sol",
            "gpt-5.6-sol-2026-07-30": "sol",
            "gpt-5.6-terra": "terra",
            "gpt-5.6-luna": "luna",
            "codex-auto-review": "review",
            "gpt-5.5": "other",
            "미분류": "other",
        }
        for model, expected in cases.items():
            with self.subTest(model=model):
                self.assertEqual(usage_report.model_label(model), expected)

    def test_models_are_comma_separated_in_fixed_order(self) -> None:
        self.assertEqual(
            usage_report.display_models({"other", "review", "sol"}),
            "sol, review, other",
        )

    def test_token_format_is_always_millions_with_two_decimals(self) -> None:
        self.assertEqual(usage_report.format_tokens(1_127_015_966), "1,127.02M")
        self.assertEqual(usage_report.format_tokens(1_000), "0.00M")

    def test_token_cell_includes_share_of_row_total(self) -> None:
        self.assertEqual(usage_report.token_cell(800_000, 1_000_000), "0.80M (80.0%)")


class ReportIntegrationTests(unittest.TestCase):
    def test_foreign_history_is_not_counted_in_root_rollout(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            sessions_root = root / "sessions"
            session_index = root / "session_index.jsonl"

            write_jsonl(session_index, [{"id": "root", "thread_name": "새 task"}])
            write_jsonl(
                sessions_root / "root.jsonl",
                [
                    {
                        "timestamp": "2026-08-03T00:00:00.000Z",
                        "type": "session_meta",
                        "payload": {
                            "id": "root",
                            "session_id": "root",
                            "cwd": "/tmp/root-project",
                            "source": "vscode",
                        },
                    },
                    {
                        "timestamp": "2026-08-03T00:00:00.001Z",
                        "type": "session_meta",
                        "payload": {"id": "old", "cwd": "/tmp/old-project"},
                    },
                    token_count(
                        "2026-08-03T00:00:00.001Z",
                        input_tokens=900_000,
                        cached_input_tokens=800_000,
                        output_tokens=10_000,
                    ),
                    {
                        "timestamp": "2026-08-03T00:00:01Z",
                        "type": "event_msg",
                        "payload": {"type": "task_started"},
                    },
                    {
                        "timestamp": "2026-08-03T00:00:02Z",
                        "type": "turn_context",
                        "payload": {
                            "cwd": "/tmp/root-project",
                            "model": "gpt-5.6-sol",
                        },
                    },
                    token_count(
                        "2026-08-03T00:00:03Z",
                        input_tokens=200_000,
                        cached_input_tokens=100_000,
                        output_tokens=10_000,
                    ),
                ],
            )

            report = usage_report.aggregate(
                sessions_root=sessions_root,
                target_date=date(2026, 8, 3),
                timezone_info=ZoneInfo("Asia/Seoul"),
                timezone_name="Asia/Seoul",
                rate_card=Path(__file__).parents[1] / "references" / "rate-card.toml",
                session_index=session_index,
                agent_memory_root=root / "agent-memory",
                computer_name="test-mac",
            )

            usage_report.assert_report_integrity(report)
            self.assertEqual(report.diagnostics.replayed_events, 1)
            self.assertEqual(report.diagnostics.aggregated_events, 1)
            self.assertEqual(report.total.total, 210_000)

    def test_inherited_parent_history_is_not_counted_in_subagent_rollout(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            sessions_root = root / "sessions"
            session_index = root / "session_index.jsonl"

            write_jsonl(
                session_index,
                [{"id": "parent", "thread_name": "부모 task"}],
            )
            write_jsonl(
                sessions_root / "child.jsonl",
                [
                    {
                        "timestamp": "2026-08-03T00:00:00.000Z",
                        "type": "session_meta",
                        "payload": {
                            "id": "child",
                            "session_id": "parent",
                            "cwd": "/tmp/child-project",
                            "parent_thread_id": "parent",
                            "source": {
                                "subagent": {
                                    "thread_spawn": {"parent_thread_id": "parent"}
                                }
                            },
                        },
                    },
                    {
                        "timestamp": "2026-08-03T00:00:00.001Z",
                        "type": "session_meta",
                        "payload": {"id": "parent", "cwd": "/tmp/parent-project"},
                    },
                    token_count(
                        "2026-08-03T00:00:00.001Z",
                        input_tokens=900_000,
                        cached_input_tokens=800_000,
                        output_tokens=10_000,
                    ),
                    {
                        "timestamp": "2026-08-03T00:00:00.002Z",
                        "type": "event_msg",
                        "payload": {"type": "task_started"},
                    },
                    {
                        "timestamp": "2026-08-03T00:00:01Z",
                        "type": "turn_context",
                        "payload": {
                            "cwd": "/tmp/child-project",
                            "model": "gpt-5.6-sol",
                        },
                    },
                    token_count(
                        "2026-08-03T00:00:02Z",
                        input_tokens=200_000,
                        cached_input_tokens=100_000,
                        output_tokens=10_000,
                    ),
                ],
            )

            report = usage_report.aggregate(
                sessions_root=sessions_root,
                target_date=date(2026, 8, 3),
                timezone_info=ZoneInfo("Asia/Seoul"),
                timezone_name="Asia/Seoul",
                rate_card=Path(__file__).parents[1] / "references" / "rate-card.toml",
                session_index=session_index,
                agent_memory_root=root / "agent-memory",
                computer_name="test-mac",
            )

            usage_report.assert_report_integrity(report)
            self.assertEqual(report.diagnostics.original_events, 2)
            self.assertEqual(report.diagnostics.replayed_events, 1)
            self.assertEqual(report.diagnostics.aggregated_events, 1)
            self.assertEqual(report.total.total, 210_000)
            self.assertIn("child-project", report.sessions)
            self.assertIn("parent", report.sessions["child-project"])

    def test_project_rows_contain_nested_titled_sessions(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            sessions_root = root / "sessions"
            session_index = root / "session_index.jsonl"

            write_jsonl(
                session_index,
                [
                    {"id": "session-a", "thread_name": "이전 제목"},
                    {"id": "session-a", "thread_name": "아즈빌 설계 검토"},
                    {"id": "session-b", "thread_name": "센서 구조 확인"},
                ],
            )
            write_jsonl(
                sessions_root / "a.jsonl",
                [
                    {
                        "timestamp": "2026-08-03T00:00:00Z",
                        "type": "session_meta",
                        "payload": {"id": "session-a", "cwd": "/tmp/azbil"},
                    },
                    {
                        "timestamp": "2026-08-03T00:01:00Z",
                        "type": "turn_context",
                        "payload": {"cwd": "/tmp/azbil", "model": "gpt-5.6-sol"},
                    },
                    token_count(
                        "2026-08-03T00:02:00Z",
                        input_tokens=1_000_000,
                        cached_input_tokens=800_000,
                        output_tokens=10_000,
                    ),
                    {
                        "timestamp": "2026-08-03T00:02:30Z",
                        "type": "event_msg",
                        "payload": {
                            "type": "token_count",
                            "info": None,
                            "rate_limits": {"primary": {"used_percent": 10}},
                        },
                    },
                    {
                        "timestamp": "2026-08-03T00:03:00Z",
                        "type": "event_msg",
                        "payload": {
                            "type": "thread_settings_applied",
                            "thread_settings": {
                                "cwd": "/tmp/azbil",
                                "model": "codex-auto-review",
                            },
                        },
                    },
                    token_count(
                        "2026-08-03T00:04:00Z",
                        input_tokens=500_000,
                        cached_input_tokens=400_000,
                        output_tokens=5_000,
                    ),
                ],
            )
            write_jsonl(
                sessions_root / "b.jsonl",
                [
                    {
                        "timestamp": "2026-08-03T01:00:00Z",
                        "type": "session_meta",
                        "payload": {"id": "session-b", "cwd": "/tmp/azbil"},
                    },
                    {
                        "timestamp": "2026-08-03T01:01:00Z",
                        "type": "turn_context",
                        "payload": {
                            "cwd": "/tmp/azbil",
                            "model": "gpt-5.6-terra",
                        },
                    },
                    token_count(
                        "2026-08-03T01:02:00Z",
                        input_tokens=200_000,
                        cached_input_tokens=100_000,
                        output_tokens=10_000,
                    ),
                ],
            )

            report = usage_report.aggregate(
                sessions_root=sessions_root,
                target_date=date(2026, 8, 3),
                timezone_info=ZoneInfo("Asia/Seoul"),
                timezone_name="Asia/Seoul",
                rate_card=Path(__file__).parents[1] / "references" / "rate-card.toml",
                session_index=session_index,
                agent_memory_root=root / "agent-memory",
                computer_name="test-mac",
            )
            usage_report.assert_report_integrity(report)
            self.assertEqual(report.diagnostics.token_events_without_usage, 1)

            markdown = usage_report.render_markdown(report)
            self.assertIn("**azbil 전체 (2개)**", markdown)
            self.assertIn("└ 아즈빌 설계 검토 | sol, review |", markdown)
            self.assertIn("└ 센서 구조 확인 | terra |", markdown)
            self.assertIn("1.73M", markdown)
            self.assertIn("**1.30M (75.4%)**", markdown)
            self.assertIn("**0.40M (23.2%)**", markdown)
            self.assertIn("**0.03M (1.4%)**", markdown)
            self.assertIn("1.20M (79.2%)", markdown)
            self.assertIn("0.30M (19.8%)", markdown)
            self.assertIn("0.01M (1.0%)", markdown)
            self.assertNotIn("sol + review", markdown)

            payload = json.loads(usage_report.report_to_json(report))
            self.assertEqual(payload["sessions"][0]["project"], "azbil")
            self.assertEqual(
                payload["sessions"][0]["sessions"][0]["models"],
                ["sol", "review"],
            )


if __name__ == "__main__":
    unittest.main()
