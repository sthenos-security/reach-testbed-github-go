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
| `reach-testbed-github-go` | Codex proof lane | Codex / pending parity replay | Successful remediation run with DB-backed clean proof | `reachable-ci-artifacts` with `release-proof/summary.json` showing `verification.clean=true` | Historical successful Codex runs exist; parity replay not yet captured in this plan | Run a fresh Codex parity lane and record run ID |
| `reach-testbed-github-go` | Claude proof lane | Claude / pending parity replay | Successful remediation run with DB-backed clean proof | `reachable-ci-artifacts` with `release-proof/summary.json` showing `verification.clean=true` | Historical successful Claude runs exist; parity replay not yet captured in this plan | Run a fresh Claude parity lane and record run ID |
| `reach-ci-github` | Copilot dispatch lane | Codex / complete | Copilot lane uses `agent_task`, not `issue_assignment` | Dispatch artifact includes `copilot-tasks.repo.db`, `copilot-dispatch.json`, and `github_task_id` | Toolkit commit `6a51c23`; Go dispatch proof run `27908810746` created task `rch_task_467353d6ed2309d6` with `github_task_id=285d791b-dd51-47eb-bf0e-3a2b65574866` | Keep using this lane for parity runs |
| `reach-testbed-github-go` | Copilot PR proof lane | Codex / complete | Copilot PR can be re-scanned and terminally verified from DB evidence | `reachable-copilot-pr-verification` artifact plus `copilot_tasks` terminal row | Verify run `27909000069` marked task `rch_task_467353d6ed2309d6` and PR `#12` as `verified` | Replay auto-trigger after next Copilot PR event |
| `reach-testbed-github-go` | Cross-agent parity comparison | Codex / implemented, awaiting fresh replay | Comparator passes for Codex, Claude, and Copilot artifacts from the same parity campaign | `agent-parity-report.json` with `ok=true` | Implemented by `ci/check-agent-parity.py` and `.github/workflows/reachable-agent-parity.yml` | Feed fresh Codex, Claude, Copilot run IDs into the parity workflow |
| `reach-testbed-github-go` | Auto-trigger Copilot verification | Codex / patched, awaiting live replay | Copilot PR event runs verifier without landing in `action_required` | Successful `Verify Copilot PR` run started by PR event | Workflow trigger moved to `pull_request_target` in commit `2f26dcf` | Wait for next Copilot PR event or replay with a new Copilot PR |

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
