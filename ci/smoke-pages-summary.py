#!/usr/bin/env python3
"""Smoke checks for public Pages summary semantics."""

from __future__ import annotations

import importlib.util
import json
import tempfile
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
MODULE_PATH = ROOT / "ci" / "build-pages-summary.py"
SANITIZE_PATH = ROOT / "ci" / "sanitize-sarif-for-upload.py"


def _load_module(path: Path, name: str):
    spec = importlib.util.spec_from_file_location(name, path)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {MODULE_PATH}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


def main() -> int:
    mod = _load_module(MODULE_PATH, "build_pages_summary")
    sanitizer = _load_module(SANITIZE_PATH, "sanitize_sarif_for_upload")
    defended_result = {
        "ruleId": "CWE/78",
        "level": "error",
        "message": {"text": "NOT EXPLOITABLE - blocked by validation"},
        "locations": [
            {
                "logicalLocation": {"name": "handler"},
                "physicalLocation": {
                    "artifactLocation": {
                        "uri": "/home/runner/work/reach-testbed-go/reach-testbed-go/internal/handlers/cwe.go"
                    },
                    "region": {"startLine": 12},
                },
            }
        ],
        "properties": {"reachabilityState": "DEFENDED", "riskLevel": "CRITICAL"},
    }
    sarif = {
        "runs": [
            {
                "tool": {"driver": {"rules": [{"id": "CWE/78", "shortDescription": {"text": "Command injection"}}]}},
                "results": [defended_result],
            }
        ]
    }
    summary = mod._summarize(sarif=sarif, ledger={}, compliance={})
    assert summary["top_priority"] == [], "defended rows must not appear in exploitable/reachable priority"
    assert len(summary["top_defended"]) == 1, "defended rows should stay in defended section"
    assert summary["top_defended"][0]["location"] == "internal/handlers/cwe.go:12"

    with tempfile.TemporaryDirectory() as tmp:
        path = Path(tmp) / "reachable-code-scanning.sarif"
        path.write_text(json.dumps(sarif), encoding="utf-8")
        data = json.loads(path.read_text(encoding="utf-8"))
        removed = sanitizer.sanitize(data)
        assert removed == 1
        assert "logicalLocation" not in data["runs"][0]["results"][0]["locations"][0]

    print("Pages summary smoke passed")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
