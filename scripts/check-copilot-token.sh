#!/usr/bin/env bash
set -euo pipefail

repo="${1:-${GITHUB_REPOSITORY:-sthenos-security/reach-testbed-github-go}}"
token="${REACHABLE_COPILOT_USER_TOKEN:-}"
api_version="${GITHUB_API_VERSION:-2026-03-10}"
api_root="${GITHUB_API_ROOT:-https://api.github.com}"

if [[ -z "$token" ]]; then
  echo "FAIL missing REACHABLE_COPILOT_USER_TOKEN" >&2
  exit 2
fi

if [[ "$repo" != */* ]]; then
  echo "FAIL repository must be owner/name, got: $repo" >&2
  exit 2
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

request() {
  local name="$1"
  local method="$2"
  local path="$3"
  local body="${4:-}"
  local out="$tmpdir/${name}.body"
  local code

  if [[ -n "$body" ]]; then
    code="$(
      curl -sS -o "$out" -w '%{http_code}' \
        -X "$method" \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $token" \
        -H "X-GitHub-Api-Version: $api_version" \
        -H "Content-Type: application/json" \
        --data "$body" \
        "$api_root$path"
    )"
  else
    code="$(
      curl -sS -o "$out" -w '%{http_code}' \
        -X "$method" \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer $token" \
        -H "X-GitHub-Api-Version: $api_version" \
        "$api_root$path"
    )"
  fi

  printf '%s\n' "$code" > "$tmpdir/${name}.code"
}

show_body() {
  local name="$1"
  python3 - "$tmpdir/${name}.body" <<'PY'
import json
import sys
from pathlib import Path

text = Path(sys.argv[1]).read_text(encoding="utf-8", errors="replace")
try:
    payload = json.loads(text)
except json.JSONDecodeError:
    print(text[:1200])
    raise SystemExit

for key in ("message", "documentation_url", "status"):
    if key in payload:
        print(f"{key}: {payload[key]}")
PY
}

owner="${repo%%/*}"

echo "Checking Copilot token for $repo"

request user GET /user
user_code="$(cat "$tmpdir/user.code")"
if [[ "$user_code" != "200" ]]; then
  echo "FAIL token is not accepted by GitHub /user (HTTP $user_code)" >&2
  show_body user >&2
  exit 1
fi
login="$(python3 - "$tmpdir/user.body" <<'PY'
import json
import sys
from pathlib import Path

print(json.loads(Path(sys.argv[1]).read_text(encoding="utf-8")).get("login", "<unknown>"))
PY
)"
echo "OK token authenticates as: $login"

request repo GET "/repos/$repo"
repo_code="$(cat "$tmpdir/repo.code")"
if [[ "$repo_code" != "200" ]]; then
  echo "FAIL token cannot read repo $repo (HTTP $repo_code)" >&2
  show_body repo >&2
  exit 1
fi
echo "OK token can read repo metadata"

request cloud_config GET "/repos/$repo/copilot/cloud-agent/configuration"
cloud_code="$(cat "$tmpdir/cloud_config.code")"
if [[ "$cloud_code" == "200" ]]; then
  echo "OK token can read repo Copilot cloud-agent configuration"
else
  echo "WARN token cannot read repo Copilot cloud-agent configuration (HTTP $cloud_code)" >&2
  show_body cloud_config >&2
fi

request agent_tasks GET "/agents/repos/$repo/tasks?per_page=1"
tasks_code="$(cat "$tmpdir/agent_tasks.code")"
if [[ "$tasks_code" == "200" ]]; then
  if python3 - "$tmpdir/agent_tasks.body" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
raise SystemExit(0 if isinstance(payload.get("tasks"), list) else 1)
PY
  then
    echo "OK token can query Copilot agent tasks for repo"
  else
    echo "FAIL Copilot agent task list response is not the expected shape" >&2
    show_body agent_tasks >&2
    exit 1
  fi
else
  echo "FAIL token cannot query Copilot agent tasks for repo (HTTP $tasks_code)" >&2
  show_body agent_tasks >&2
  exit 1
fi

request org_permissions GET "/orgs/$owner/copilot/coding-agent/permissions"
org_code="$(cat "$tmpdir/org_permissions.code")"
if [[ "$org_code" == "200" ]]; then
  echo "OK token can read org Copilot Coding Agent permissions"
  python3 - "$tmpdir/org_permissions.body" <<'PY'
import json
import sys
from pathlib import Path

payload = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
for key in ("enabled_repositories", "selected_repositories_url"):
    if key in payload:
        print(f"{key}: {payload[key]}")
PY
else
  echo "WARN token cannot read org Copilot Coding Agent permissions (HTTP $org_code)"
  show_body org_permissions
fi

echo "Non-mutating token checks passed."
echo "If reachctl copilot dispatch still fails with HTTP 409, GitHub accepted the token but has not enabled Copilot Coding Agent task creation for this user/repo/org."
