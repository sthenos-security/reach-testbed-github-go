#!/usr/bin/env python3
"""Write sanitized remediation audit evidence for agent proof artifacts."""

from __future__ import annotations

import argparse
import json
import os
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def _load_json(path: Path | None) -> dict[str, Any]:
    if path is None or not path.exists():
        return {}
    raw = path.read_text(encoding="utf-8")
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        decoder = json.JSONDecoder()
        for index, char in enumerate(raw):
            if char != "{":
                continue
            try:
                parsed, _ = decoder.raw_decode(raw[index:])
            except json.JSONDecodeError:
                continue
            if isinstance(parsed, dict):
                return parsed
    return {"_error": f"failed to parse {path}"}


def _read_lines(path: Path | None) -> list[str]:
    if path is None or not path.exists():
        return []
    return [line.strip() for line in path.read_text(encoding="utf-8").splitlines() if line.strip()]


def _workflow_context() -> dict[str, Any]:
    repository = os.environ.get("GITHUB_REPOSITORY", "")
    run_id = os.environ.get("GITHUB_RUN_ID", "")
    server_url = os.environ.get("GITHUB_SERVER_URL", "https://github.com").rstrip("/")
    return {
        "provider": "github_actions" if os.environ.get("GITHUB_ACTIONS") == "true" else "local",
        "repository": repository,
        "workflow": os.environ.get("GITHUB_WORKFLOW", ""),
        "job": os.environ.get("GITHUB_JOB", ""),
        "run_id": run_id,
        "run_attempt": os.environ.get("GITHUB_RUN_ATTEMPT", ""),
        "run_url": f"{server_url}/{repository}/actions/runs/{run_id}" if repository and run_id else "",
        "event_name": os.environ.get("GITHUB_EVENT_NAME", ""),
        "actor": os.environ.get("GITHUB_ACTOR", ""),
        "ref": os.environ.get("GITHUB_REF", ""),
        "sha": os.environ.get("GITHUB_SHA", ""),
    }


def _proof_totals(summary: dict[str, Any]) -> dict[str, int]:
    totals = summary.get("totals") if isinstance(summary.get("totals"), dict) else {}
    keys = [
        "signals",
        "release_blockers",
        "reachable",
        "supply_chain",
        "manual_review",
        "malware_alerts",
        "copilot_tasks",
    ]
    return {key: int(totals.get(key) or 0) for key in keys}


def _copilot_task(summary: dict[str, Any]) -> dict[str, Any]:
    tasks = summary.get("copilot_tasks") if isinstance(summary.get("copilot_tasks"), list) else []
    if not tasks or not isinstance(tasks[0], dict):
        return {}
    task = tasks[0]
    return {
        "task_id": str(task.get("task_id") or ""),
        "state": str(task.get("state") or ""),
        "verification_status": str(task.get("verification_status") or ""),
        "selected_signal_ids": list(task.get("selected_signal_ids") or []),
        "selected_files": list(task.get("selected_files") or []),
        "remediation_key": str(task.get("remediation_key") or ""),
        "pr_url": str(task.get("pr_url") or ""),
    }


def _write_markdown(path: Path, audit: dict[str, Any]) -> None:
    if audit.get("kind") == "parity":
        _write_parity_markdown(path, audit)
        return

    totals = audit.get("proof", {}).get("totals", {})
    changed_files = audit.get("changed_files") or []
    task = audit.get("copilot_task") or {}
    lines = [
        "# Agent Remediation Audit Log",
        "",
        f"- Created: `{audit['created_at']}`",
        f"- Lane: `{audit['lane']}`",
        f"- Repository: `{audit['repository']}`",
        f"- Branch: `{audit['branch']}`",
        f"- Pull request: `{audit['pull_request'].get('url') or audit['pull_request'].get('number') or 'n/a'}`",
        f"- Run: `{audit['run'].get('url') or audit['run'].get('id') or 'n/a'}`",
        "",
        "## Verdict",
        "",
        f"- Status: **{audit['outcome']['status']}**",
        f"- Message: {audit['outcome']['message']}",
        f"- Release blockers after fix: `{totals.get('release_blockers', 0)}`",
        f"- Reachable findings after fix: `{totals.get('reachable', 0)}`",
        "",
        "## Copilot Task",
        "",
        f"- Task ID: `{task.get('task_id') or 'n/a'}`",
        f"- State: `{task.get('state') or 'n/a'}`",
        f"- Verification: `{task.get('verification_status') or 'n/a'}`",
        f"- Selected signals: `{len(task.get('selected_signal_ids') or [])}`",
        "",
        "## Changed Files",
        "",
    ]
    if changed_files:
        lines.extend(f"- `{name}`" for name in changed_files)
    else:
        lines.append("No changed-file list was captured.")
    lines.extend(
        [
            "",
            "## Build/Test Evidence",
            "",
            f"- Go test log captured: `{str(audit.get('test_evidence', {}).get('go_test_log_present', False)).lower()}`",
        ]
    )
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def _write_parity_markdown(path: Path, audit: dict[str, Any]) -> None:
    lines = [
        "# Agent Parity Audit Log",
        "",
        f"- Created: `{audit['created_at']}`",
        f"- Status: **{audit['outcome']['status']}**",
        f"- Message: {audit['outcome']['message']}",
        "",
        "| Lane | Clean | Blocking results | Branch |",
        "|---|---:|---:|---|",
    ]
    lanes = audit.get("lanes") if isinstance(audit.get("lanes"), dict) else {}
    for lane_name in ("codex", "claude", "copilot"):
        lane = lanes.get(lane_name) if isinstance(lanes.get(lane_name), dict) else {}
        lines.append(
            f"| `{lane_name}` | `{str(lane.get('clean', False)).lower()}` | "
            f"`{lane.get('blocking_results', '')}` | `{lane.get('branch', '')}` |"
        )
    mismatches = audit.get("mismatches") or []
    lines.extend(["", "## Mismatches", ""])
    if mismatches:
        lines.extend(f"- {mismatch}" for mismatch in mismatches)
    else:
        lines.append("No mismatches.")
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")


def _write_outputs(output_dir: Path, audit: dict[str, Any]) -> None:
    output_dir.mkdir(parents=True, exist_ok=True)
    json_path = output_dir / "agent-remediation-audit-log.json"
    md_path = output_dir / "agent-remediation-audit-log.md"
    json_path.write_text(json.dumps(audit, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    _write_markdown(md_path, audit)
    print(f"Agent audit log written: {json_path}")
    print(f"Agent audit log written: {md_path}")


def _write_pr_audit(args: argparse.Namespace) -> int:
    summary = _load_json(args.proof_summary)
    verify = _load_json(args.verify_json)
    totals = _proof_totals(summary)
    task = _copilot_task(summary)
    changed_files = _read_lines(args.changed_files)
    result = str(verify.get("result") or "")
    task_verified = result == "verified" and task.get("verification_status") == "verified"
    clean = totals.get("release_blockers", 0) == 0
    status = "verified_clean" if task_verified and clean else "insufficient_coverage"
    message = (
        "Copilot task is verified and the post-fix proof is clean."
        if status == "verified_clean"
        else "Copilot task evidence exists, but the post-fix proof is not clean enough for parity."
    )
    audit = {
        "schema_version": 1,
        "kind": "agent_pr",
        "created_at": datetime.now(timezone.utc).isoformat(),
        "lane": args.lane,
        "repository": args.repository or summary.get("repository") or os.environ.get("GITHUB_REPOSITORY", ""),
        "branch": args.branch or summary.get("branch") or "",
        "commit": summary.get("commit") or "",
        "workflow": _workflow_context(),
        "pull_request": {"number": args.pr_number or "", "url": args.pr_url or task.get("pr_url") or ""},
        "run": {"id": os.environ.get("GITHUB_RUN_ID", ""), "url": args.run_url or ""},
        "proof": {
            "summary_path": str(args.proof_summary) if args.proof_summary else "",
            "totals": totals,
            "clean": clean,
        },
        "copilot_task": task,
        "verification": {
            "result": result,
            "absent_signal_ids": (verify.get("evaluation") or {}).get("absent_signal_ids", [])
            if isinstance(verify.get("evaluation"), dict)
            else [],
            "json_path": str(args.verify_json) if args.verify_json else "",
        },
        "changed_files": changed_files,
        "changed_file_count": len(changed_files),
        "test_evidence": {
            "go_test_log": str(args.go_test_log) if args.go_test_log else "",
            "go_test_log_present": bool(args.go_test_log and args.go_test_log.exists()),
        },
        "outcome": {"status": status, "message": message},
    }
    _write_outputs(args.output_dir, audit)
    return 0


def _write_parity_audit(args: argparse.Namespace) -> int:
    report = _load_json(args.parity_report)
    mismatches = list(report.get("mismatches") or [])
    ok = bool(report.get("ok"))
    audit = {
        "schema_version": 1,
        "kind": "parity",
        "created_at": datetime.now(timezone.utc).isoformat(),
        "parity_report": str(args.parity_report),
        "lanes": report.get("lanes") or {},
        "mismatches": mismatches,
        "outcome": {
            "status": "verified_clean" if ok else "insufficient_coverage",
            "message": "All agent lanes reached the same clean outcome."
            if ok
            else "At least one agent lane did not reach clean parity.",
        },
    }
    _write_outputs(args.output_dir, audit)
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--lane", default="copilot-github")
    parser.add_argument("--artifact-dir", type=Path, default=Path(".reachable/ci-artifacts"))
    parser.add_argument("--output-dir", type=Path, default=None)
    parser.add_argument("--proof-summary", type=Path, default=None)
    parser.add_argument("--verify-json", type=Path, default=None)
    parser.add_argument("--changed-files", type=Path, default=None)
    parser.add_argument("--go-test-log", type=Path, default=None)
    parser.add_argument("--pr-number", default="")
    parser.add_argument("--pr-url", default="")
    parser.add_argument("--run-url", default="")
    parser.add_argument("--repository", default="")
    parser.add_argument("--branch", default="")
    parser.add_argument("--parity-report", type=Path, default=None)
    args = parser.parse_args()
    if args.output_dir is None:
        args.output_dir = args.artifact_dir
    if args.parity_report is not None:
        return _write_parity_audit(args)
    return _write_pr_audit(args)


if __name__ == "__main__":
    raise SystemExit(main())
