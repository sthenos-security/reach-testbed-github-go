#!/usr/bin/env python3
"""Compare Codex, Claude, and Copilot proof artifacts by security outcome."""

from __future__ import annotations

import argparse
import json
import sqlite3
import sys
from pathlib import Path
from typing import Any


def _find_one(root: Path, pattern: str) -> Path:
    matches = sorted(root.rglob(pattern))
    if not matches:
        raise FileNotFoundError(f"missing {pattern} under {root}")
    return matches[0]


def _find_optional(root: Path, pattern: str) -> Path | None:
    matches = sorted(root.rglob(pattern))
    return matches[0] if matches else None


def _load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def _load_json_with_prefix_noise(path: Path) -> dict[str, Any]:
    raw = path.read_text(encoding="utf-8")
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
    raise json.JSONDecodeError("no JSON object found", raw, 0)


def _require_str(value: Any, *, field: str) -> str:
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"missing or invalid string field: {field}")
    return value


def _normalize_code_agent(name: str, root: Path) -> dict[str, Any]:
    summary_path = _find_one(root, "summary.json")
    summary = _load_json(summary_path)
    verification = summary.get("verification") if isinstance(summary.get("verification"), dict) else {}
    totals = summary.get("totals") if isinstance(summary.get("totals"), dict) else {}
    repo = str(summary.get("repo") or summary.get("repository") or "").strip()
    if not repo:
        raise ValueError(f"missing repo/repository field in {summary_path}")
    if verification:
        status = str(verification.get("status") or "")
        clean = bool(verification.get("clean"))
        blocking_results = int(verification.get("blocking_results") or 0)
        branch = str(verification.get("branch") or summary.get("ref") or summary.get("branch") or "")
        run_id = str(summary.get("run_id") or "")
    else:
        blocking_results = int(totals.get("release_blockers") or 0)
        clean = blocking_results == 0
        status = "legacy_release_proof" if clean else "legacy_blocking"
        branch = str(summary.get("branch") or "")
        run_url = str(summary.get("run_url") or "")
        run_id = run_url.rstrip("/").split("/")[-1] if run_url else ""
    return {
        "lane": name,
        "artifact_root": str(root),
        "summary_path": str(summary_path),
        "repository": repo,
        "branch": branch,
        "run_id": run_id,
        "status": status,
        "clean": clean,
        "blocking_results": blocking_results,
        "ok": clean and blocking_results == 0,
    }


def _copilot_task_row(db_path: Path) -> dict[str, Any]:
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    try:
        row = conn.execute(
            """
            SELECT task_id, state, verification_status, pr_number, pr_url, verification_run_url
            FROM copilot_tasks
            ORDER BY updated_at DESC, rowid DESC
            LIMIT 1
            """
        ).fetchone()
    finally:
        conn.close()
    if row is None:
        raise ValueError(f"no copilot_tasks row found in {db_path}")
    return dict(row)


def _normalize_copilot(root: Path) -> dict[str, Any]:
    verify_path = _find_one(root, "copilot-verify-pr.json")
    proof_summary_path = _find_one(root, "summary.json")
    verified_db_path = _find_one(root, "copilot-verified-tasks.repo.db")
    go_test_log_path = _find_optional(root, "copilot-go-test.log")
    audit_log_path = _find_optional(root, "agent-remediation-audit-log.json")

    verify = _load_json_with_prefix_noise(verify_path)
    proof_summary = _load_json(proof_summary_path)
    task_row = _copilot_task_row(verified_db_path)

    repository = _require_str(proof_summary.get("repository"), field="copilot.repository")
    copilot_tasks = proof_summary.get("copilot_tasks") or []
    if not isinstance(copilot_tasks, list) or not copilot_tasks:
        raise ValueError(f"missing copilot_tasks in {proof_summary_path}")
    task = copilot_tasks[0]

    result = str(verify.get("result") or "")
    evaluation = verify.get("evaluation") if isinstance(verify.get("evaluation"), dict) else {}
    verification = (
        proof_summary.get("verification")
        if isinstance(proof_summary.get("verification"), dict)
        else {}
    )
    totals = proof_summary.get("totals") if isinstance(proof_summary.get("totals"), dict) else {}
    if verification:
        proof_status = str(verification.get("status") or "")
        clean = bool(verification.get("clean"))
        blocking_results = int(verification.get("blocking_results") or 0)
    else:
        blocking_results = int(totals.get("release_blockers") or 0)
        clean = blocking_results == 0
        proof_status = "legacy_release_proof" if clean else "legacy_blocking"
    task_verified = result == "verified" and str(task_row.get("verification_status") or "") == "verified"
    return {
        "lane": "copilot",
        "artifact_root": str(root),
        "verify_path": str(verify_path),
        "summary_path": str(proof_summary_path),
        "verified_db_path": str(verified_db_path),
        "go_test_log_path": str(go_test_log_path) if go_test_log_path else "",
        "has_go_test_log": go_test_log_path is not None,
        "audit_log_path": str(audit_log_path) if audit_log_path else "",
        "has_audit_log": audit_log_path is not None,
        "repository": repository,
        "branch": str(proof_summary.get("branch") or ""),
        "task_id": str(task_row.get("task_id") or task.get("task_id") or ""),
        "state": str(task_row.get("state") or task.get("state") or ""),
        "verification_status": str(task_row.get("verification_status") or task.get("verification_status") or ""),
        "pr_number": task_row.get("pr_number"),
        "pr_url": str(task_row.get("pr_url") or task.get("pr_url") or verify.get("pr_url") or ""),
        "verification_run_url": str(task_row.get("verification_run_url") or verify.get("run_url") or ""),
        "result": result,
        "status": proof_status,
        "clean": clean,
        "blocking_results": blocking_results,
        "absent_signal_ids": list(evaluation.get("absent_signal_ids") or []),
        "task_verified": task_verified,
        "ok": task_verified and clean and blocking_results == 0,
    }


def _compare_agents(
    codex: dict[str, Any],
    claude: dict[str, Any],
    copilot: dict[str, Any],
    *,
    expected_repository: str | None,
    expected_copilot_task_id: str | None,
    expected_copilot_pr: int | None,
) -> dict[str, Any]:
    mismatches: list[str] = []
    repositories = {codex["repository"], claude["repository"], copilot["repository"]}
    if len(repositories) != 1:
        mismatches.append(
            "repository mismatch across lanes: "
            + ", ".join(f"{lane['lane']}={lane['repository']}" for lane in (codex, claude, copilot))
        )
    if expected_repository and repositories != {expected_repository}:
        mismatches.append(f"expected repository {expected_repository}, saw {sorted(repositories)}")
    if not codex["ok"]:
        mismatches.append(
            f"codex proof is not clean: status={codex['status']} clean={codex['clean']} "
            f"blocking_results={codex['blocking_results']}"
        )
    if not claude["ok"]:
        mismatches.append(
            f"claude proof is not clean: status={claude['status']} clean={claude['clean']} "
            f"blocking_results={claude['blocking_results']}"
        )
    if not copilot["ok"]:
        mismatches.append(
            f"copilot proof is not clean and verified: state={copilot['state']} "
            f"verification_status={copilot['verification_status']} result={copilot['result']} "
            f"status={copilot['status']} clean={copilot['clean']} "
            f"blocking_results={copilot['blocking_results']}"
        )
    if not copilot["has_go_test_log"]:
        mismatches.append("copilot proof is missing copilot-go-test.log")
    if not copilot["has_audit_log"]:
        mismatches.append("copilot proof is missing agent-remediation-audit-log.json")
    if expected_copilot_task_id and copilot["task_id"] != expected_copilot_task_id:
        mismatches.append(
            f"expected Copilot task {expected_copilot_task_id}, saw {copilot['task_id']}"
        )
    if expected_copilot_pr is not None and copilot["pr_number"] != expected_copilot_pr:
        mismatches.append(f"expected Copilot PR #{expected_copilot_pr}, saw {copilot['pr_number']}")
    return {
        "ok": not mismatches,
        "mismatches": mismatches,
        "lanes": {
            "codex": codex,
            "claude": claude,
            "copilot": copilot,
        },
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--codex-artifact-dir", type=Path, required=True)
    parser.add_argument("--claude-artifact-dir", type=Path, required=True)
    parser.add_argument("--copilot-artifact-dir", type=Path, required=True)
    parser.add_argument("--expected-repository", default=None)
    parser.add_argument("--expected-copilot-task-id", default=None)
    parser.add_argument("--expected-copilot-pr", type=int, default=None)
    parser.add_argument("--output", type=Path, default=None)
    args = parser.parse_args()

    codex = _normalize_code_agent("codex", args.codex_artifact_dir)
    claude = _normalize_code_agent("claude", args.claude_artifact_dir)
    copilot = _normalize_copilot(args.copilot_artifact_dir)
    report = _compare_agents(
        codex,
        claude,
        copilot,
        expected_repository=args.expected_repository,
        expected_copilot_task_id=args.expected_copilot_task_id,
        expected_copilot_pr=args.expected_copilot_pr,
    )

    payload = json.dumps(report, indent=2, sort_keys=True)
    if args.output is not None:
        args.output.write_text(payload + "\n", encoding="utf-8")
    print(payload)
    return 0 if report["ok"] else 1


if __name__ == "__main__":
    raise SystemExit(main())
