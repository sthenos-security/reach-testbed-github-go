# Copilot Parity Test Plan

This document is the active parity tracker for `reach-testbed-github-go`.

## Goal

Prove that the Go demo repository can produce the same security outcome across
the three supported remediation lanes:

- `openai-codex`
- `anthropic-claude`
- `copilot-github`

Parity is measured at the proof level, not at the diff-text level. The patches
may differ; the validated security outcome must match.

## Acceptance Criteria

All of the following must be true before the parity goal is considered complete:

1. The same repository and target baseline are used for all three lanes.
2. Codex remediation produces a DB-backed clean proof artifact.
3. Claude remediation produces a DB-backed clean proof artifact.
4. Copilot dispatch produces a DB-backed task artifact with a real
   `github_task_id`.
5. Copilot opens a PR for the dispatched task.
6. `reachctl copilot verify-pr` marks the Copilot task `verified` from a
   post-fix scan DB.
7. A parity comparison run reports that all three lanes reached the same
   verified security outcome.

## Tracker

| Repo | Task | Owner / Status | Acceptance Criteria | Proof Required | Latest Evidence | Next Action |
|---|---|---|---|---|---|---|
| `reach-testbed-github-go` | Codex proof lane | Codex / complete | Successful remediation run with DB-backed clean proof | `reachable-ci-artifacts` with `release-proof/summary.json` showing zero release blockers | Fresh parity replay run `27912011654` succeeded; artifact `7777713855`; PR `#17`; `release_blockers=0`, `reachable=0` | Keep artifact for parity comparison |
| `reach-testbed-github-go` | Claude proof lane | Claude / complete | Successful remediation run with DB-backed clean proof | `reachable-ci-artifacts` with `release-proof/summary.json` showing zero release blockers | Fresh parity replay run `27912011701` succeeded; artifact `7777726968`; PR `#18`; `release_blockers=0`, `reachable=0` | Keep artifact for parity comparison |
| `reach-ci-github` | Copilot dispatch lane | Codex / complete | Copilot lane uses `agent_task`, not `issue_assignment` | Dispatch artifact includes `copilot-tasks.repo.db`, `copilot-dispatch.json`, and `github_task_id` | Fresh Go dispatch proof run `27912011678` created task `rch_task_467353d6ed2309d6` with `github_task_id=791c6b88-8b88-45e2-9272-5b74bba49571` | Use PR `#16` for DB-backed verification |
| `reach-testbed-github-go` | Copilot PR proof lane | Codex / complete | Copilot PR can be re-scanned and terminally verified from DB evidence | `reachable-copilot-pr-verification` artifact plus `copilot_tasks` terminal row | Fresh verification run `27912272154` succeeded; artifact `7777723020`; task `rch_task_467353d6ed2309d6`, PR `#16`, `verification_status=verified` | Keep artifact for parity comparison |
| `reach-testbed-github-go` | Cross-agent parity comparison | Codex / complete | Comparator passes for Codex, Claude, and Copilot artifacts from the same parity campaign | `agent-parity-report.json` with `ok=true` | Remote parity workflow run `27912378610` succeeded; artifact `7777743207`; report has `ok=true` and `mismatches=[]` | Keep this campaign as the current acceptance proof |
| `reach-testbed-github-go` | Trusted Copilot verification dispatcher | Codex / implemented, awaiting live replay | A trusted `main` workflow finds Copilot PRs and dispatches `Verify Copilot PR` without loosening public PR approval policy | Successful `Dispatch Copilot PR Verification` run that starts `Verify Copilot PR` for an `app/copilot-swe-agent` PR | PR-event runs for PR `#16` were blocked by GitHub as `action_required`; dispatcher workflow now avoids that gate by using trusted `workflow_dispatch` from `main` | Run dispatcher against PR `#16` with `force=true`, then inspect verifier run and labels |

## Live Campaign - 2026-06-21

| Lane | Workflow | Run ID | Status | Required Proof | Next Action |
|---|---|---:|---|---|---|
| Codex | `Run Demo (Codex)` | `27912011654` | Complete | `reachable-ci-artifacts/release-proof/summary.json` clean DB-backed proof | Artifact `7777713855`; PR `#17`; `release_blockers=0`, `reachable=0` |
| Claude | `Run Demo (Claude)` | `27912011701` | Complete | `reachable-ci-artifacts/release-proof/summary.json` clean DB-backed proof | Artifact `7777726968`; PR `#18`; `release_blockers=0`, `reachable=0` |
| Copilot dispatch | `Run Demo (Copilot Dispatch)` | `27912011678` | Complete | `copilot-dispatch.json` with `dispatch_kind=agent_task` plus `copilot-tasks.repo.db` | Fresh artifact proves task `rch_task_467353d6ed2309d6`, `dispatch_kind=agent_task`, `github_task_id=791c6b88-8b88-45e2-9272-5b74bba49571` |
| Copilot cloud agent | `Copilot cloud agent` | `27912164940` | Complete | Successful Copilot task processing and PR authored by `app/copilot-swe-agent` | Opened PR `#16`, branch `copilot/rch-task-467353d6ed2309d6-remediate-dlp-again` |
| Copilot verification | `Verify Copilot PR` | `27912272154` | Complete | `reachable-copilot-pr-verification` with task row `verification_status=verified` | Artifact `7777723020`; task `rch_task_467353d6ed2309d6`, PR `#16`, result `verified` |
| Parity comparison | `Agent Parity Check` | `27912378610` | Complete | `reachable-agent-parity-report/agent-parity-report.json` with `ok=true` | Artifact `7777743207`; report has `ok=true`, `mismatches=[]`, Codex run `27912011654`, Claude run `27912011701`, Copilot verification run `27912272154` |

## Required Run Set

The parity campaign should use the following sequence:

1. Run `Run Demo (Codex)` against `main`.
2. Run `Run Demo (Claude)` against `main`.
3. Run `Run Demo (Copilot Dispatch)` against `main`.
4. Wait for the corresponding `Copilot cloud agent` PR.
5. Run `Verify Copilot PR` against that PR and dispatch run.
6. Run `Agent Parity Check` with the four resulting run IDs.

## Artifact Contract

The parity workflow compares proof artifacts, not commits:

- Codex / Claude:
  - `reachable-ci-artifacts`
  - `release-proof/summary.json`
- Copilot dispatch:
  - `reachable-ci-artifacts`
  - `copilot-dispatch.json`
  - `copilot-tasks.repo.db`
- Copilot verification:
  - `reachable-copilot-pr-verification`
  - `copilot-verify-pr.json`
  - `copilot-verified-tasks.repo.db`
  - `release-proof/summary.json`

## Comparison Rules

The parity workflow must pass only when:

1. All artifacts are from the same repository.
2. Codex proof is clean.
3. Claude proof is clean.
4. Copilot verification status is `verified`.
5. Optional expected Copilot task / PR inputs, when provided, match the proof.

The workflow does not require identical:

- diff text
- commit history
- branch names
- PR titles
- internal coding style

## Current Known Good Copilot Evidence

- Dispatch run: `27908810746`
- Copilot task: `rch_task_467353d6ed2309d6`
- GitHub task ID: `285d791b-dd51-47eb-bf0e-3a2b65574866`
- Copilot PR: `#12`
- Verification run: `27909000069`
- Verification result: `verified`
