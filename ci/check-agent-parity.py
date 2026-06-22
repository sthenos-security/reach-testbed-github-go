#!/usr/bin/env python3
"""Compare Codex, Claude, and Copilot proof artifacts by security outcome."""

from __future__ import annotations

import argparse
import json
import sqlite3
import sys
from collections import defaultdict
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


def _copilot_task_rows(db_path: Path) -> list[dict[str, Any]]:
    conn = sqlite3.connect(db_path)
    conn.row_factory = sqlite3.Row
    try:
        rows = conn.execute(
            """
            SELECT
                task_id,
                state,
                verification_status,
                pr_number,
                pr_url,
                verification_run_url,
                selected_signal_ids_json
            FROM copilot_tasks
            ORDER BY rowid ASC
            """
        ).fetchall()
    finally:
        conn.close()
    parsed: list[dict[str, Any]] = []
    for row in rows:
        raw_ids = row["selected_signal_ids_json"]
        try:
            selected_signal_ids = json.loads(raw_ids) if raw_ids else []
        except json.JSONDecodeError:
            selected_signal_ids = []
        parsed.append(
            {
                "task_id": str(row["task_id"] or ""),
                "state": str(row["state"] or ""),
                "verification_status": str(row["verification_status"] or ""),
                "pr_number": row["pr_number"],
                "pr_url": str(row["pr_url"] or ""),
                "verification_run_url": str(row["verification_run_url"] or ""),
                "selected_signal_ids": [
                    str(value) for value in selected_signal_ids if isinstance(value, str) and value.strip()
                ],
            }
        )
    if not parsed:
        raise ValueError(f"no copilot_tasks row found in {db_path}")
    return parsed


def _normalize_copilot_artifact(root: Path) -> dict[str, Any]:
    verify_path = _find_one(root, "copilot-verify-pr.json")
    proof_summary_path = _find_one(root, "summary.json")
    verified_db_path = _find_one(root, "copilot-verified-tasks.repo.db")
    go_test_log_path = _find_optional(root, "copilot-go-test.log")
    audit_log_path = _find_optional(root, "agent-remediation-audit-log.json")

    verify = _load_json_with_prefix_noise(verify_path)
    proof_summary = _load_json(proof_summary_path)
    task_rows = _copilot_task_rows(verified_db_path)
    task_rows_by_id = {str(row.get("task_id") or ""): row for row in task_rows}
    raw_task_id = str(verify.get("task_id") or "")
    if raw_task_id:
        task_row = task_rows_by_id.get(raw_task_id)
    else:
        task_row = None
    if task_row is None:
        task_row = _copilot_task_row(verified_db_path)
        raw_task_id = str(task_row.get("task_id") or raw_task_id)

    repository = _require_str(proof_summary.get("repository"), field="copilot.repository")
    copilot_tasks = proof_summary.get("copilot_tasks") or []
    if not isinstance(copilot_tasks, list) or not copilot_tasks:
        raise ValueError(f"missing copilot_tasks in {proof_summary_path}")
    task = next(
        (
            candidate
            for candidate in copilot_tasks
            if str(candidate.get("task_id") or "") == raw_task_id
        ),
        copilot_tasks[0],
    )

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
        "kind": "copilot_artifact",
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
        "task_id": raw_task_id or str(task_row.get("task_id") or task.get("task_id") or ""),
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
        "defended_signal_ids": list(evaluation.get("defended_signal_ids") or []),
        "still_vulnerable_signal_ids": list(evaluation.get("still_vulnerable_signal_ids") or []),
        "needs_review_signal_ids": list(evaluation.get("needs_review_signal_ids") or []),
        "task_verified": task_verified,
        "ok": task_verified and clean and blocking_results == 0,
        "task_rows": task_rows,
    }


def _copilot_artifact_roots(root: Path) -> list[Path]:
    direct = root / "copilot-verify-pr.json"
    if direct.is_file():
        return [root]
    matches = sorted({path.parent for path in root.rglob("copilot-verify-pr.json")})
    if not matches:
        raise FileNotFoundError(f"missing copilot-verify-pr.json under {root}")
    return matches


def _copilot_status_rank(status: str) -> int:
    return {
        "verified": 5,
        "still_vulnerable": 4,
        "needs_human_review": 3,
        "error": 2,
        "verification_running": 1,
        "pending": 0,
        "": 0,
    }.get(status, 0)


def _merge_task_rows(artifacts: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    merged: dict[str, dict[str, Any]] = {}
    for artifact in artifacts:
        for row in artifact.get("task_rows") or []:
            task_id = str(row.get("task_id") or "")
            if not task_id:
                continue
            current = merged.get(task_id)
            if current is None:
                merged[task_id] = dict(row)
                continue
            current_rank = _copilot_status_rank(str(current.get("verification_status") or ""))
            candidate_rank = _copilot_status_rank(str(row.get("verification_status") or ""))
            if candidate_rank > current_rank:
                merged[task_id] = dict(row)
                continue
            if candidate_rank == current_rank:
                if not current.get("pr_url") and row.get("pr_url"):
                    merged[task_id] = dict(row)
                    continue
                if not current.get("verification_run_url") and row.get("verification_run_url"):
                    merged[task_id] = dict(row)
                    continue
    return merged


def _normalize_copilot(root: Path) -> dict[str, Any]:
    artifact_roots = _copilot_artifact_roots(root)
    artifacts = [_normalize_copilot_artifact(path) for path in artifact_roots]
    repositories = {artifact["repository"] for artifact in artifacts}
    if len(repositories) != 1:
        raise ValueError(f"repository mismatch across Copilot artifacts under {root}: {sorted(repositories)}")
    task_rows = _merge_task_rows(artifacts)
    if not task_rows:
        raise ValueError(f"no Copilot task rows found under {root}")

    selected_signal_ids: set[str] = set()
    for row in task_rows.values():
        for signal_id in row.get("selected_signal_ids") or []:
            if isinstance(signal_id, str) and signal_id.strip():
                selected_signal_ids.add(signal_id)

    verified_task_ids = {
        task_id
        for task_id, row in task_rows.items()
        if str(row.get("verification_status") or "") == "verified"
    }
    pending_task_ids = sorted(set(task_rows) - verified_task_ids)

    resolved_signal_ids: set[str] = set()
    still_vulnerable_signal_ids: set[str] = set()
    needs_review_signal_ids: set[str] = set()
    artifact_task_ids: dict[str, list[str]] = defaultdict(list)
    verified_pr_numbers: set[int] = set()
    verified_pr_urls: set[str] = set()
    missing_go_test_logs: list[str] = []
    missing_audit_logs: list[str] = []

    for artifact in artifacts:
        task_id = str(artifact.get("task_id") or "")
        if task_id:
            artifact_task_ids[task_id].append(str(artifact.get("artifact_root") or ""))
        if artifact.get("task_verified"):
            if not artifact.get("has_go_test_log"):
                missing_go_test_logs.append(task_id or str(artifact.get("artifact_root") or ""))
            if not artifact.get("has_audit_log"):
                missing_audit_logs.append(task_id or str(artifact.get("artifact_root") or ""))
            resolved_signal_ids.update(str(signal_id) for signal_id in artifact.get("absent_signal_ids") or [])
            resolved_signal_ids.update(str(signal_id) for signal_id in artifact.get("defended_signal_ids") or [])
            still_vulnerable_signal_ids.update(
                str(signal_id) for signal_id in artifact.get("still_vulnerable_signal_ids") or []
            )
            needs_review_signal_ids.update(
                str(signal_id) for signal_id in artifact.get("needs_review_signal_ids") or []
            )
            pr_number = artifact.get("pr_number")
            if isinstance(pr_number, int):
                verified_pr_numbers.add(pr_number)
            pr_url = str(artifact.get("pr_url") or "")
            if pr_url:
                verified_pr_urls.add(pr_url)

    unresolved_signal_ids = sorted(selected_signal_ids - resolved_signal_ids)
    coverage_clean = (
        not unresolved_signal_ids
        and not still_vulnerable_signal_ids
        and not needs_review_signal_ids
    )
    task_verified = bool(task_rows) and len(verified_task_ids) == len(task_rows)
    all_have_go_test_log = not missing_go_test_logs
    all_have_audit_log = not missing_audit_logs

    blocking_results = (
        len(unresolved_signal_ids)
        + len(still_vulnerable_signal_ids)
        + len(needs_review_signal_ids)
    )
    branch = "campaign" if len(artifacts) > 1 or len(task_rows) > 1 else artifacts[0]["branch"]
    primary_task_id = next(iter(verified_task_ids), "") if len(verified_task_ids) == 1 else ""
    primary_pr_number = next(iter(verified_pr_numbers), None) if len(verified_pr_numbers) == 1 else None
    primary_pr_url = next(iter(verified_pr_urls), "") if len(verified_pr_urls) == 1 else ""

    return {
        "kind": "copilot_campaign",
        "lane": "copilot",
        "artifact_root": str(root),
        "artifact_roots": [str(path) for path in artifact_roots],
        "repository": next(iter(repositories)),
        "branch": branch,
        "task_id": primary_task_id,
        "task_ids": sorted(task_rows),
        "state": "campaign_verified" if task_verified else "campaign_incomplete",
        "verification_status": "verified" if task_verified else "incomplete",
        "pr_number": primary_pr_number,
        "pr_numbers": sorted(verified_pr_numbers),
        "pr_url": primary_pr_url,
        "pr_urls": sorted(verified_pr_urls),
        "result": "verified" if coverage_clean and task_verified else "still_vulnerable",
        "status": "campaign_verified" if coverage_clean and task_verified else "campaign_incomplete",
        "clean": coverage_clean,
        "blocking_results": blocking_results,
        "task_verified": task_verified,
        "ok": task_verified and coverage_clean and all_have_go_test_log and all_have_audit_log,
        "has_go_test_log": all_have_go_test_log,
        "has_audit_log": all_have_audit_log,
        "missing_go_test_logs": missing_go_test_logs,
        "missing_audit_logs": missing_audit_logs,
        "selected_signal_ids": sorted(selected_signal_ids),
        "resolved_signal_ids": sorted(resolved_signal_ids),
        "unresolved_signal_ids": unresolved_signal_ids,
        "still_vulnerable_signal_ids": sorted(still_vulnerable_signal_ids),
        "needs_review_signal_ids": sorted(needs_review_signal_ids),
        "verified_task_ids": sorted(verified_task_ids),
        "pending_task_ids": pending_task_ids,
        "task_rows": [task_rows[task_id] for task_id in sorted(task_rows)],
        "artifacts": artifacts,
        "artifact_task_ids": dict(artifact_task_ids),
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
            f"copilot proof is not campaign-complete: state={copilot['state']} "
            f"verification_status={copilot['verification_status']} result={copilot['result']} "
            f"status={copilot['status']} clean={copilot['clean']} "
            f"blocking_results={copilot['blocking_results']} "
            f"verified_tasks={len(copilot.get('verified_task_ids') or [])}/"
            f"{len(copilot.get('task_ids') or [])}"
        )
    if not copilot["has_go_test_log"]:
        missing = ", ".join(copilot.get("missing_go_test_logs") or [])
        mismatches.append(f"copilot proof is missing copilot-go-test.log for: {missing or '<unknown>'}")
    if not copilot["has_audit_log"]:
        missing = ", ".join(copilot.get("missing_audit_logs") or [])
        mismatches.append(
            f"copilot proof is missing agent-remediation-audit-log.json for: {missing or '<unknown>'}"
        )
    if expected_copilot_task_id and expected_copilot_task_id not in set(copilot.get("verified_task_ids") or []):
        mismatches.append(
            f"expected Copilot task {expected_copilot_task_id}, saw {copilot.get('verified_task_ids') or []}"
        )
    if expected_copilot_pr is not None and expected_copilot_pr not in set(copilot.get("pr_numbers") or []):
        mismatches.append(f"expected Copilot PR #{expected_copilot_pr}, saw {copilot.get('pr_numbers') or []}")
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
