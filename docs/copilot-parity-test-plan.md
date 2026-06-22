# Copilot Parity Test Plan

This document is the active design note and parity tracker for
`reach-testbed-github-go`.

## Goal

Prove that the Go demo repository can produce the same security outcome across
the three supported remediation lanes:

- `openai-codex`
- `anthropic-claude`
- `copilot-github`

Parity is measured at the proof level, not at the diff-text level. The patches
may differ; the validated security outcome must match. The project goal is not
"Copilot opened a PR" or "one Copilot task verified"; the goal is high-quality
Copilot remediation with comparable coverage to Codex and Claude on the same
baseline.

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
7. Copilot's post-fix proof artifact shows zero release blockers on the same
   scope that Codex and Claude cleared.
8. Copilot's PR verification runs the repository build/test gate and publishes
   the test log.
9. Copilot's PR verification publishes a sanitized remediation audit log
   artifact that records the task, selected signals, changed files, test
   evidence, and post-fix blocker counts.
10. A parity comparison run reports that all three lanes reached the same
    verified security outcome, whether Copilot proves that through one PR or a
    verified multi-PR campaign.
11. Agent prompts explicitly require mergeable PRs: scoped diffs, no broad
    formatting sweeps, no unrelated renames/refactors, no unselected dependency
    churn, and clear blocker reporting when a mergeable bounded fix is not
    possible.

## Tracker

| Repo | Task | Owner / Status | Acceptance Criteria | Proof Required | Latest Evidence | Next Action |
|---|---|---|---|---|---|---|
| `reach-testbed-github-go` | Codex proof lane | Codex / complete | Successful remediation run with DB-backed clean proof | `reachable-ci-artifacts` with `release-proof/summary.json` showing zero release blockers | Fresh parity replay run `27912011654` succeeded; artifact `7777713855`; PR `#17`; `release_blockers=0`, `reachable=0` | Keep artifact for parity comparison |
| `reach-testbed-github-go` | Claude proof lane | Claude / complete | Successful remediation run with DB-backed clean proof | `reachable-ci-artifacts` with `release-proof/summary.json` showing zero release blockers | Fresh parity replay run `27912011701` succeeded; artifact `7777726968`; PR `#18`; `release_blockers=0`, `reachable=0` | Keep artifact for parity comparison |
| `reach-ci-github` | Copilot dispatch lane | Codex / complete | Copilot lane uses `agent_task`, not `issue_assignment` | Dispatch artifact includes `copilot-tasks.repo.db`, `copilot-dispatch.json`, and `github_task_id` | Fresh Go dispatch proof run `27912011678` created task `rch_task_467353d6ed2309d6` with `github_task_id=791c6b88-8b88-45e2-9272-5b74bba49571` | Use PR `#16` for DB-backed verification |
| `reach-ci-github` | Copilot fixed-input replay artifact | Codex / in progress | Copilot dispatch artifacts preserve the exact selected bundle plus a pre-dispatch repo DB for cheap replay | `copilot-bundle.json`, `copilot-prompt.md`, and `copilot-pre-dispatch.repo.db` inside `reachable-ci-artifacts` | Artifact contract patched locally; pending first live replay run from a fresh dispatch built with the updated workflow | Re-run `Run Demo (Copilot Dispatch)` once, then use `Run Demo (Copilot Bundle Replay)` for fast task-shaping iterations |
| `reach-testbed-github-go` | Copilot PR proof lane | Codex / insufficient | Copilot PR can be re-scanned, build-tested, audit-logged, and terminally verified from DB evidence | `reachable-copilot-pr-verification` artifact plus `copilot-go-test.log`, `agent-remediation-audit-log.json`, `release-proof/summary.json`, and `copilot_tasks` terminal row | Verification run `27912272154` proved task `rch_task_467353d6ed2309d6` on PR `#16`, but stricter local replay reports `blocking_results=17`; the PR only changed `internal/handlers/dlp.go` and did not prove full remediation coverage | Re-run after the verifier includes build/test evidence, audit-log evidence, and post-fix clean proof |
| `reach-testbed-github-go` | Copilot high-quality remediation coverage | Codex / open | Copilot clears the same actionable release-blocker scope as Codex and Claude, including dependency/module-version remediation where required | Copilot post-fix `release-proof/summary.json` with zero release blockers, plus `agent-remediation-audit-log.json` showing selected signals and changed files | Current Copilot PR `#16` is narrow DLP-only; Codex PR `#17` and Claude PR `#18` addressed broader handler files and `go.mod`/`go.sum` | Broaden Copilot dispatch/campaign until coverage matches Codex and Claude outcomes |
| `reach-testbed-github-go` | Cross-agent parity comparison | Codex / in progress | Comparator passes for Codex, Claude, and Copilot artifacts from the same parity campaign, with Copilot accepted either as one clean PR or as a fully verified multi-PR campaign that covers the same selected signals | `agent-parity-report.json` and aggregate `agent-remediation-audit-log.json` with `ok=true`, zero unresolved Copilot selected signals, and no mismatches | Run `27912378610` passed under the old narrow rule; local synthetic replay now proves the updated comparator can treat two verified Copilot PR artifacts as one clean campaign when their verified signal coverage resolves the full selected-signal set | Re-run parity once the live sharded Copilot campaign produces real verification artifacts |
| `reach-testbed-github-go` | Trusted Copilot verification dispatcher | Codex / complete | A trusted `main` workflow finds Copilot PRs and dispatches `Verify Copilot PR` without loosening public PR approval policy | Successful `Dispatch Copilot PR Verification` run and successful spawned verifier for an `app/copilot-swe-agent` PR | Dispatcher run `27913235873` started verifier run `27913239123`; verifier succeeded and artifact `7778016209` has `verification_status=verified`; fixed dispatcher run `27913328451` succeeded and skipped PR `#16` because it found that verification artifact | Keep strict PR-event approval policy; use dispatcher for Copilot PR verification automation |
| `reach-core` | Mergeable remediation prompt contract | Codex / in progress | Codex, Claude, and Copilot prompts instruct agents to produce mergeable scoped PRs and avoid cross-shard conflict churn | Prompt text and tests proving mergeability instructions are present | Local patch adds mergeability guidance to shared remediation prompts and Copilot issue/task bodies | Run focused prompt tests and include the fix in the next alpha build |
| `reach-testbed-github-go` | Copilot merge policy documentation | Codex / complete | README documents multi-PR campaign mode, manual trusted merge, and auto-merge as an opt-in roadmap feature | `docs/copilot-remediation-campaign.md` plus README link | Added campaign doc with PRs `#22`-`#26`; live GitHub PR metadata reports all five are `MERGEABLE` with `mergeStateStatus=UNSTABLE` until verification/status checks run | Verify each PR before any merge |

## Merge Policy

The first product cut does not require Copilot to auto-merge. Copilot creates
one PR per REACHABLE shard. REACHABLE verifies each PR, and maintainers may
merge the verified PRs themselves according to normal repository policy.

Manual maintainer merge is acceptable for this demo after the proof gates pass:

1. `Verify Copilot PR` succeeds for the PR.
2. `copilot-verify-pr.json` reports `verified`.
3. The post-fix proof has zero selected blockers for that task.
4. `go test ./...` evidence is present.
5. The campaign parity report passes against Codex and Claude.

Auto-merge is a roadmap feature, only for customers that request it and grant a
trusted REACHABLE workflow or GitHub App the required repository permission.
GitHub does not provide a native source-branch-only merge permission for
`copilot/rch-task-*`; the control must be REACHABLE policy gates plus normal
branch protection. Any future automation must use the verified PR head SHA via
`--match-head-commit` to prevent post-verification drift.

## Live Campaign - 2026-06-21

| Lane | Workflow | Run ID | Status | Required Proof | Next Action |
|---|---|---:|---|---|---|
| Codex | `Run Demo (Codex)` | `27912011654` | Complete | `reachable-ci-artifacts/release-proof/summary.json` clean DB-backed proof | Artifact `7777713855`; PR `#17`; `release_blockers=0`, `reachable=0` |
| Claude | `Run Demo (Claude)` | `27912011701` | Complete | `reachable-ci-artifacts/release-proof/summary.json` clean DB-backed proof | Artifact `7777726968`; PR `#18`; `release_blockers=0`, `reachable=0` |
| Copilot dispatch | `Run Demo (Copilot Dispatch)` | `27912011678` | Complete | `copilot-dispatch.json` with `dispatch_kind=agent_task` plus `copilot-tasks.repo.db` | Fresh artifact proves task `rch_task_467353d6ed2309d6`, `dispatch_kind=agent_task`, `github_task_id=791c6b88-8b88-45e2-9272-5b74bba49571` |
| Copilot cloud agent | `Copilot cloud agent` | `27912164940` | Complete | Successful Copilot task processing and PR authored by `app/copilot-swe-agent` | Opened PR `#16`, branch `copilot/rch-task-467353d6ed2309d6-remediate-dlp-again` |
| Copilot verification | `Verify Copilot PR` | `27912272154` | Insufficient | `reachable-copilot-pr-verification` with task row `verification_status=verified`, build/test log, audit log, and clean post-fix proof | Artifact `7777723020`; task `rch_task_467353d6ed2309d6`, PR `#16`, result `verified`; coverage was DLP-only |
| Parity comparison | `Agent Parity Check` | `27912378610` | Insufficient | `reachable-agent-parity-report/agent-parity-report.json` and aggregate audit log with `ok=true` under the strict clean-Copilot rule | Artifact `7777743207` passed the old rule; that proof is retired as a final parity claim |

## Current Copilot Shard Contract

The current Copilot campaign should not dispatch one 13-rule mixed batch. The
current live dry run against alpha-proof artifact `27960935439` and the updated
`reach-core` sharder produces five tasks:

1. `internal/handlers/cwe.go` — 3 rules (`EXPLOITABLE_NOW` plus same-file
   follow-on flow rules)
2. `internal/handlers/ai.go`, `internal/handlers/cve.go`,
   `internal/handlers/suspicious.go` — 4 rules (fetch/error-disclosure shard)
3. `internal/handlers/cve.go` + package target `golang.org/x/text` — 1 rule
   (`DEPENDENCY_SWEEP`)
4. `internal/handlers/ai.go`, `internal/handlers/dlp.go` — 4 rules
   (`REACHABLE_CRITICAL` outbound sensitive-data shard)
5. `internal/handlers/secrets.go` — 1 rule (`BACKLOG` secret cleanup)

This is the active task shape to prove next. The fresh `reach-core` candidate
build for that proof was triggered as run `27973261407` on branch
`agent-plugin-alpha-proof`.

The live alpha candidate run `27978630290` proved this five-task shape and
Copilot opened PRs `#22` through `#26`. GitHub currently reports all five PRs
as `MERGEABLE`; they are not product-ready until REACHABLE verification and the
campaign parity check pass.

## Shard Design Rules

Copilot sharding should follow the same remediation ordering used for the other
agent lanes instead of inventing a Copilot-only batching strategy:

1. Order by priority lane first: `EXPLOITABLE_NOW`, then reachable critical
   source findings, then dependency sweeps, then lower-priority cleanup.
2. Preserve affinity groups. If one package upgrade or one shared remediation
   action clears multiple findings, keep that work in the same task.
3. Keep same-file follow-on findings attached to the earlier higher-priority
   shard for that file when the fix path is shared.
4. Keep dependency-only upgrades separate from source-code exploit-path shards
   unless shard-count overflow forces a small adjacent merge.
5. Use the shard count to cap task size, not to flatten priority. The highest
   exploitability work should still be dispatched first.

## Required Run Set

The parity campaign should use the following sequence:

1. Run `Run Demo (Codex)` against `main`.
2. Run `Run Demo (Claude)` against `main`.
3. Run `Run Demo (Copilot Dispatch)` against `main`.
4. If the first Copilot result is weak, run `Run Demo (Copilot Bundle Replay)`
   against the saved dispatch artifact instead of rescanning the repo.
5. Wait for the corresponding `Copilot cloud agent` PR.
6. Run `Verify Copilot PR` against that PR and dispatch run.
7. If Copilot only fixes a subset of the release blockers, dispatch additional
   Copilot tasks or a broader Copilot campaign until the post-fix proof is
   clean.
8. Run `Agent Parity Check` with the final clean run IDs.

## Artifact Contract

The parity workflow compares proof artifacts, not commits:

- Codex / Claude:
  - `reachable-ci-artifacts`
  - `release-proof/summary.json`
- Copilot dispatch:
  - `reachable-ci-artifacts`
  - `copilot-dispatch.json`
  - `copilot-tasks.repo.db`
  - `copilot-pre-dispatch.repo.db`
  - `copilot-bundle.json`
  - `copilot-prompt.md`
- Copilot verification:
  - `reachable-copilot-pr-verification`
  - `copilot-go-test.log`
  - `copilot-dispatch.json`
  - `agent-remediation-audit-log.json`
  - `agent-remediation-audit-log.md`
  - `copilot-verify-pr.json`
  - `copilot-verified-tasks.repo.db`
  - `release-proof/summary.json`
- Parity comparison:
  - `reachable-agent-parity-report`
  - `agent-parity-report.json`
  - `agent-remediation-audit-log.json`
  - `agent-remediation-audit-log.md`

## Comparison Rules

The parity workflow must pass only when:

1. All artifacts are from the same repository.
2. Codex proof is clean.
3. Claude proof is clean.
4. Copilot verification status is `verified`.
5. Copilot post-fix proof is clean with zero release blockers.
6. Copilot verification includes `copilot-go-test.log`.
7. Copilot verification includes `agent-remediation-audit-log.json`.
8. Optional expected Copilot task / PR inputs, when provided, match the proof.

The workflow does not require identical:

- diff text
- commit history
- branch names
- PR titles
- internal coding style

## Historical Narrow Copilot Evidence

- Dispatch run: `27908810746`
- Copilot task: `rch_task_467353d6ed2309d6`
- GitHub task ID: `285d791b-dd51-47eb-bf0e-3a2b65574866`
- Copilot PR: `#12`
- Verification run: `27909000069`
- Verification result: `verified`

This evidence is useful for proving the asynchronous Copilot task flow, but it
is not sufficient for the current parity goal because it does not prove full
remediation coverage.
