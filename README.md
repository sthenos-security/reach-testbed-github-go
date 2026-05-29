# reach-testbed-go

Intentionally vulnerable Go fixture repository for exercising REACHABLE signal
families against a compact `net/http` service.

> Do not deploy this application. It contains synthetic security issues for
> scanner validation only.

## Implementation Plan

1. Keep the fixture Go-centric and dependency-light: one `cmd/server` entrypoint
   with handlers in `internal/handlers`.
2. Model one reachable case for each major signal family:
   CVE, CWE, secret, config, DLP, AI/LLM misuse, malware/suspicious behavior.
3. Include comparison cases for `UNKNOWN` / assess, defended findings, and a
   no-fix CVE with a documented compensating control.
4. Track dependency-upgrade expectations in `go.mod`, `go.sum`, and the
   customer-facing baseline manifest in `EXPECTED.md` so scanner regressions
   can be diffed without guessing.
5. Keep all secrets and DLP data synthetic.

## Layout

```text
cmd/server/              HTTP entrypoint and route registration
internal/handlers/       Reachable, defended, and assess signal cases
internal/safety/         Small guard helpers used by defended cases
config/                  Synthetic insecure configuration cases
deploy/                  Synthetic IaC cases for config scanners
testdata/dlp/            Synthetic DLP corpus
EXPECTED.md              Customer-facing baseline manifest of expected findings
```

## Local Smoke Test

```bash
go test ./...
go run ./cmd/server
```

The service listens on `:8080` by default.

## GitHub Actions Remediation Demo

This repository is the compact Go demo for the Reachable CI remediation story:

1. Reachable scans the target branch and writes `repo.db`.
2. Reachable generates `.reachable/remediation-bundle/` from database-backed
   signal truth.
3. A selected coding agent applies one bounded remediation batch.
4. CI runs `go test`, rescans, audits, and proves the branch is clean.
5. The workflow pushes one reviewable branch named
   `reachable-remediate-<run-id>`.

The scanner and proof steps are the same for every agent. Only the executor
step changes.

### Workflow Configuration

Recommended manual inputs:

```yaml
workflow_dispatch:
  inputs:
    agent:
      description: "Coding agent executor"
      type: choice
      default: claude
      options:
        - claude
        - codex
        - opencode
    remediate:
      description: "Allow CI to create code changes"
      type: boolean
      default: false
    rescan_only:
      description: "Only prove an existing remediation branch"
      type: boolean
      default: false
    target_branch:
      description: "Base branch, or existing remediation branch for rescan_only"
      type: string
      default: main
    max_batches:
      description: "Maximum serialized remediation batches"
      type: number
      default: 1
```

Recommended secrets:

```yaml
env:
  REACHABLE_API_KEY: ${{ secrets.REACHABLE_API_KEY }}
  GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  MCP_GITHUB_TOKEN: ${{ secrets.MCP_GITHUB_TOKEN }}

  OPENROUTER_API_KEY: ${{ secrets.OPENROUTER_API_KEY }}
  ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
  OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
  GROQ_API_KEY: ${{ secrets.GROQ_API_KEY }}
  DEEPSEEK_API_KEY: ${{ secrets.DEEPSEEK_API_KEY }}
  MOONSHOT_API_KEY: ${{ secrets.MOONSHOT_API_KEY }}

  CODEX_API_KEY: ${{ secrets.CODEX_API_KEY }}
  CLAUDE_CODE_OAUTH_TOKEN: ${{ secrets.CLAUDE_CODE_OAUTH_TOKEN }}
  OPENCODE_AUTH: ${{ secrets.OPENCODE_AUTH }}
```

Token roles:

- `GITHUB_TOKEN`: branch, commit, PR, SARIF/code-scanning, and GitHub REST
  access.
- `MCP_GITHUB_TOKEN`: fine-grained GitHub token for MCP-based agent access.
- `REACHABLE_API_KEY`: optional Reachable cloud publish / org attach.
- LLM keys: Reachable scan/enrichment providers and, depending on the selected
  agent action, coding-agent model provider auth.

### Job Shape

```yaml
permissions:
  contents: write
  pull-requests: write
  security-events: write

jobs:
  reachable-remediate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ inputs.target_branch }}

      - name: Install Reachable
        run: |
          curl -fsSL https://get.reachable.security/install.sh | bash -s -- --ci --no-vibe
          echo "$HOME/.reachable/venv/bin" >> "$GITHUB_PATH"

      - name: Baseline scan
        run: reachctl scan . --ci --sarif reachable.sarif

      - name: Stop if remediation is disabled
        if: ${{ !inputs.remediate || inputs.rescan_only }}
        run: |
          go test ./...
          reachctl audit --latest --summary
          reachctl integrity --latest

      - name: Create remediation branch
        if: ${{ inputs.remediate && !inputs.rescan_only }}
        run: |
          BRANCH="reachable-remediate-${GITHUB_RUN_ID}"
          git switch -c "$BRANCH"
          echo "REACHABLE_REMEDIATION_BRANCH=$BRANCH" >> "$GITHUB_ENV"

      - name: Generate Reachable prompt bundle
        if: ${{ inputs.remediate && !inputs.rescan_only }}
        run: |
          reachctl vibe prompt \
            --workspace . \
            --agent "${{ inputs.agent }}" \
            --branch-name "$REACHABLE_REMEDIATION_BRANCH" \
            --all

      - name: Install selected coding agent
        if: ${{ inputs.remediate && !inputs.rescan_only }}
        run: |
          case "${{ inputs.agent }}" in
            claude) npm install -g @anthropic-ai/claude-code ;;
            codex) npm install -g @openai/codex ;;
            opencode) npm install -g opencode-ai ;;
          esac

      - name: Run selected coding agent
        if: ${{ inputs.remediate && !inputs.rescan_only }}
        run: ./ci/run-agent.sh "${{ inputs.agent }}" .reachable/remediation-bundle/prompt.md

      - name: Prove remediation
        if: ${{ inputs.remediate || inputs.rescan_only }}
        run: |
          go test ./...
          reachctl scan . --ci --sarif reachable-after.sarif
          reachctl audit --latest --summary
          reachctl integrity --latest
```

`remediate=false` is the safe default. It proves the scanner output without
allowing CI to edit code. Set `remediate=true` only when you want the workflow
to create a remediation branch and run the selected coding agent.

For larger repositories, run serialized batches:

```text
scan -> select batch -> write prompt/audit -> run agent -> test -> rescan -> next batch
```

Do not send an unbounded backlog to a single prompt.

The checked-in workflow lives at
`.github/workflows/reachable-remediate.yml`; the only agent-specific shim is
`ci/run-agent.sh`.
