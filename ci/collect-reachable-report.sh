#!/usr/bin/env bash
set -euo pipefail

label="${1:-scan}"
out_root="${2:-.reachable/ci-artifacts}"
scan_path="${SCAN_PATH:-}"

if [ -z "$scan_path" ]; then
  scan_path="$(python - <<'PY'
from pathlib import Path

root = Path.home() / ".reachable" / "scans"
candidates = [p for p in root.glob("*/*/20*") if p.is_dir()]
print(max(candidates, key=lambda p: p.stat().st_mtime) if candidates else "")
PY
)"
fi

if [ -z "$scan_path" ] || [ ! -d "$scan_path" ]; then
  echo "No Reachable scan session found under ~/.reachable/scans." >&2
  exit 2
fi

safe_label="$(printf '%s' "$label" | tr -c 'A-Za-z0-9_.-' '-')"
report_dir="$out_root/reports/$safe_label"
mkdir -p "$report_dir"

printf '%s\n' "$scan_path" > "$report_dir/scan-path.txt"

if [ -f "$scan_path/scan.log" ]; then
  cp "$scan_path/scan.log" "$report_dir/scan.log"
fi

if [ -f "$scan_path/prof.json" ]; then
  cp "$scan_path/prof.json" "$report_dir/prof.json"
fi

set +e
reachctl audit --scan-path "$scan_path" --summary --verbose > "$report_dir/audit.txt" 2>&1
audit_status="$?"
reachctl integrity --scan-path "$scan_path" > "$report_dir/integrity.txt" 2>&1
integrity_status="$?"
if reachctl compliance --help >/dev/null 2>&1; then
  reachctl compliance report --scan-path "$scan_path" --format markdown --output "$report_dir/compliance.md" > "$report_dir/compliance.stdout" 2> "$report_dir/compliance.stderr"
  compliance_md_status="$?"
  reachctl compliance report --scan-path "$scan_path" --format json --output "$report_dir/compliance.json" >> "$report_dir/compliance.stdout" 2>> "$report_dir/compliance.stderr"
  compliance_json_status="$?"
  if reachctl compliance narrative --help >/dev/null 2>&1; then
    reachctl compliance narrative --scan-path "$scan_path" --dry-run --format markdown --output "$report_dir/compliance-narrative.md" >> "$report_dir/compliance.stdout" 2>> "$report_dir/compliance.stderr"
    compliance_narrative_md_status="$?"
    reachctl compliance narrative --scan-path "$scan_path" --dry-run --format json --output "$report_dir/compliance-narrative.json" >> "$report_dir/compliance.stdout" 2>> "$report_dir/compliance.stderr"
    compliance_narrative_json_status="$?"
  else
    compliance_narrative_md_status="127"
    compliance_narrative_json_status="127"
  fi
else
  printf '%s\n' "reachctl compliance command is not available in this wheel." > "$report_dir/compliance.stderr"
  compliance_md_status="127"
  compliance_json_status="127"
  compliance_narrative_md_status="127"
  compliance_narrative_json_status="127"
fi
set -e

printf '%s\n' "$audit_status" > "$report_dir/audit.exitcode"
printf '%s\n' "$integrity_status" > "$report_dir/integrity.exitcode"
printf '%s\n' "$compliance_md_status" > "$report_dir/compliance-md.exitcode"
printf '%s\n' "$compliance_json_status" > "$report_dir/compliance-json.exitcode"
printf '%s\n' "$compliance_narrative_md_status" > "$report_dir/compliance-narrative-md.exitcode"
printf '%s\n' "$compliance_narrative_json_status" > "$report_dir/compliance-narrative-json.exitcode"

cat > "$report_dir/README.md" <<EOF
# Reachable Scan Report: $safe_label

- Scan path: \`$scan_path\`
- Audit exit code: \`$audit_status\`
- Integrity exit code: \`$integrity_status\`

Issue-bearing artifacts:

- \`../*.sarif\` - machine-readable issue report for GitHub code scanning
- \`scan.log\` - scan console log
- \`audit.txt\` - data-quality and issue audit
- \`integrity.txt\` - SARIF/database integrity proof
- \`compliance.md\` / \`compliance.json\` - DB-backed compliance evidence pack when supported by the installed wheel
- \`compliance-narrative.md\` / \`compliance-narrative.json\` - evidence-cited auditor narrative draft when supported by the installed wheel
EOF

echo "Reachable scan report artifact prepared: $report_dir"
