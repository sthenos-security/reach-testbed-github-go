# Reachable Compliance Evidence Pack

- Source: `repo.db` compliance tables
- Framework: `All frameworks`
- Report ID: `2`
- Scan ID: `4`
- Fingerprint: `1a5fc732e4a391d2e9c0693838c379b8d860508dd554ad319c922c223b9df661`
- Generated: `2026-06-04 18:50:21`

## Scan

- Branch: `reachable-remediate-26972297438`
- Commit: `8abd7831`
- Timestamp: `2026-06-04 18:50:16`
- Status: `complete`

## Summary

| Metric | Count |
| --- | ---: |
| Total signals | 8 |
| Actionable production signals | 0 |
| Reachable/Exploitable production signals | 0 |
| Unknown production signals | 0 |
| Non-production signals | 7 |
| Not reachable signals | 8 |
| Remediation units | 0 |

## Risk Distribution

| Risk | Count |
| --- | ---: |
| `CRITICAL` | 0 |
| `HIGH` | 0 |
| `MEDIUM` | 0 |
| `LOW` | 1 |
| `INFO` | 7 |
| `NONE` | 0 |

## Evidence

| Evidence ID | Status | Framework | Control | Severity | Location(s) | Evidence |
| ---: | --- | --- | --- | --- | --- | --- |
| `53` | `PASS` | Reachable evidence | `REV-SCAN-001` | `HIGH` | repo.db | Scan produced 8 canonical signal rows in repo.db. |
| `60` | `PASS` | Reachable evidence | `REV-AI-BOM-001` | `MEDIUM` | repo.db | 0 deterministic model/RAG/MCP/tool output flow(s) are materialized as AI-BOM entries for governance reporting. |
| `57` | `PASS` | Reachable evidence | `REV-FEEDBACK-001` | `MEDIUM` | repo.db | 0 feedback outcome(s) are tracked; 0 require human review and 0 are policy waivers. |
| `55` | `PASS` | Reachable evidence | `REV-PROD-001` | `MEDIUM` | repo.db | 7 signal(s) are marked non-production and excluded from actionable CI issue transport. |
| `58` | `PASS` | Reachable evidence | `REV-PROOF-001` | `MEDIUM` | repo.db | 0 optional proof prompt(s) are tracked across remediation runs; 0 cover MCP/RAG or agentic runtime-supply-chain risk. |
| `59` | `PASS` | Reachable evidence | `REV-PROOF-RUN-001` | `MEDIUM` | repo.db | 0 proof run(s) are recorded across 0 profile(s); 0 verified exploitable and 0 defended after re-attack. |
| `54` | `PASS` | Reachable evidence | `REV-REACH-001` | `MEDIUM` | repo.db | 0 production signal(s) remain reachable or unknown; 0 are reachable/exploitable. |
| `56` | `PASS` | Reachable evidence | `REV-REMED-001` | `MEDIUM` | repo.db | 0 remediation unit(s) are materialized for agent prompt generation. |
| `61` | `PASS` | SLSA | `SLSA-Dependencies` | `MEDIUM` | repo.db | 0 dependency/package risk signal(s) are present in the scan. |

## How To Read This Evidence

This pack is generated from the same scan database used by the demo verdict.
The important auditor fields are the scan ID, branch, commit, timestamp,
control mapping, evidence status, and source location. The CI run also
publishes the matching summary, remediation verdict, integrity proof, and
platform export so the evidence can be traced back to one remediation branch
and one proof scan.

This evidence pack is an operational security report, not a legal attestation.
