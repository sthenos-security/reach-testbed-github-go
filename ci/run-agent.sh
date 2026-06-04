#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat >&2 <<'EOF'
Usage: ci/run-agent.sh AGENT PROMPT_PATH

Supported AGENT values:
  claude    Run Claude Code non-interactively.
  codex     Run Codex CLI non-interactively.
  opencode  Run OpenCode non-interactively when installed.
  custom    Run REACHABLE_AGENT_RUN_COMMAND with PROMPT_PATH in the environment.

The script is intentionally thin. Reachable owns scan, bundle generation,
audit artifacts, and proof. The selected coding agent only consumes prompt.md
and edits the current branch.
EOF
}

if [[ "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ $# -ne 2 ]]; then
  usage
  exit 2
fi

AGENT="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]')"
PROMPT_PATH="$2"

if [[ ! -f "$PROMPT_PATH" ]]; then
  echo "prompt file not found: $PROMPT_PATH" >&2
  exit 2
fi

case "$AGENT" in
  claude)
    command -v claude >/dev/null 2>&1 || {
      echo "claude CLI not found. Install Claude Code or select another agent." >&2
      exit 127
    }
    claude \
      --print \
      --permission-mode bypassPermissions \
      --no-session-persistence \
      --output-format stream-json \
      --max-budget-usd "${CLAUDE_MAX_BUDGET_USD:-5}" \
      < "$PROMPT_PATH"
    ;;

  codex)
    command -v codex >/dev/null 2>&1 || {
      echo "codex CLI not found. Install Codex or select another agent." >&2
      exit 127
    }
    codex exec \
      --dangerously-bypass-approvals-and-sandbox \
      -C "$PWD" \
      --skip-git-repo-check \
      --ephemeral \
      - < "$PROMPT_PATH"
    ;;

  opencode)
    command -v opencode >/dev/null 2>&1 || {
      echo "opencode CLI not found. Install OpenCode or select another agent." >&2
      exit 127
    }
    opencode_args=(run --dir "$PWD" --dangerously-skip-permissions)
    if [[ -n "${OPENCODE_MODEL:-}" ]]; then
      opencode_args+=(--model "$OPENCODE_MODEL")
    fi
    if [[ -n "${OPENCODE_AGENT:-}" ]]; then
      opencode_args+=(--agent "$OPENCODE_AGENT")
    fi
    opencode "${opencode_args[@]}" --file "$PROMPT_PATH" "Apply the Reachable remediation instructions in the attached prompt file."
    ;;

  custom)
    if [[ -z "${REACHABLE_AGENT_RUN_COMMAND:-}" ]]; then
      echo "REACHABLE_AGENT_RUN_COMMAND is required when AGENT=custom." >&2
      exit 2
    fi
    PROMPT_PATH="$PROMPT_PATH" bash -lc "$REACHABLE_AGENT_RUN_COMMAND"
    ;;

  *)
    echo "unsupported agent: $AGENT" >&2
    usage
    exit 2
    ;;
esac
