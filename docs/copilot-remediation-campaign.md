# Copilot Remediation Campaign

This document describes the supported first product cut for GitHub Copilot
cloud-agent remediation in this Go demo repository.

## Product Model

Copilot is an asynchronous GitHub remediation lane. REACHABLE does not run
Copilot as a local CLI process. Instead:

1. REACHABLE scans the target branch and builds a DB-backed remediation bundle.
2. REACHABLE shards the selected findings by priority and remediation affinity.
3. Each shard becomes one bounded GitHub Copilot cloud-agent task.
4. Copilot opens one pull request per task.
5. REACHABLE verifies each Copilot PR from a post-fix scan database.
6. The parity workflow aggregates verified Copilot PRs as one campaign and
   compares the outcome with Codex and Claude.

The product claim is not that Copilot creates one monolithic PR. The product
claim is that REACHABLE turns prioritized risk into bounded Copilot tasks,
verifies each PR, and produces campaign-level proof.

## Multiple PRs Are Expected

Copilot campaign mode can create multiple PRs from one dispatch run. This is
intentional. Each PR corresponds to one REACHABLE remediation shard, not to the
whole scan.

REACHABLE shards findings so that Copilot gets bounded work:

- exploitable and release-blocking findings come first;
- findings that share one dependency upgrade stay together;
- findings that need coordinated edits in the same file or handler family stay
  together;
- dependency-only changes stay separate from source-code exploit-path changes
  unless REACHABLE explicitly groups them;
- unrelated risky edits are split so a weak Copilot result does not block the
  whole campaign.

The campaign is complete only when all required Copilot PRs pass
`Verify Copilot PR` and the aggregate `Agent Parity Check` shows the same
security outcome as the Codex and Claude lanes.

## Current Campaign Shape

The current Go campaign dispatches five Copilot tasks:

| Shard | Scope | Expected PR |
|---|---|---|
| 1 | Command injection and same-file follow-on rules in `internal/handlers/cwe.go` | One source-code PR |
| 2 | SSRF, cleartext fetch, tainted URL flow, and error disclosure across handler files | One source-code PR |
| 3 | `golang.org/x/text` CVE dependency upgrade | One dependency PR |
| 4 | LLM input isolation and DLP/PII redaction | One source-code PR |
| 5 | Hardcoded token / secret cleanup | One source-code PR |

The latest alpha candidate dispatch run `27978630290` produced five tasks and
Copilot opened PRs `#22` through `#26` from branches matching
`copilot/rch-task-*`.

## Merge Policy

For this demo repository, maintainers may merge the Copilot PRs manually after
REACHABLE verification passes. That is different from asking Copilot to
auto-merge its own work.

The supported v1 policy is:

1. Copilot opens PRs only.
2. `Verify Copilot PR` runs `go test ./...`, performs a post-fix REACHABLE
   scan, records `reachctl copilot verify-pr`, and uploads proof artifacts.
3. `Agent Parity Check` proves the campaign outcome against Codex and Claude.
4. A maintainer merges verified PRs according to normal repository policy.

If maintainers merge manually, use the PR head SHA visible in GitHub or via
`gh pr view --json headRefOid` as the review anchor. A trusted automation path
should use `gh pr merge --match-head-commit` so the PR cannot change after
verification and before merge.

## Mergeability Requirement For Agents

All remediation prompts must ask agents to produce mergeable PRs:

- keep changes scoped to selected findings;
- avoid broad formatting sweeps;
- avoid unrelated renames or refactors;
- update only selected dependency targets;
- keep source-code shards and dependency-only shards separate unless REACHABLE
  grouped them;
- report a blocker instead of widening scope when a safe mergeable fix is not
  possible.

This applies to Codex, Claude, and Copilot. A PR being Git-mergeable is not
enough for product proof; the PR still needs REACHABLE verification and parity
evidence.

## Auto-Merge Roadmap

Auto-merge is not a v1 default. It is a roadmap feature for customers who ask
for it and are willing to grant a trusted REACHABLE workflow or GitHub App the
required repository permissions.

The intended opt-in design is:

1. Require an explicit policy switch such as a `reachable-automerge` label.
2. Require PR author `app/copilot-swe-agent`.
3. Require head branch `copilot/rch-task-*`.
4. Require the task ID to exist in the dispatch artifact.
5. Require `copilot-verify-pr.json` result `verified`.
6. Require clean post-fix proof with zero selected blockers.
7. Require project tests and required status checks to pass.
8. Merge with `--match-head-commit` to prevent post-verification drift.

GitHub does not provide a native permission that says a token may merge only
source branches matching `copilot/rch-task-*`. The safe control is therefore
policy gating inside REACHABLE plus normal branch protection, not giving
Copilot broad merge rights.
