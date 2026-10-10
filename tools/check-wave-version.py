#!/usr/bin/env python3
"""Fail before building with a compiler other than the repository's pin."""
from pathlib import Path
import re
import subprocess
import sys

expected = (Path(__file__).resolve().parents[1] / 'wave-version').read_text().strip()
compiler = sys.argv[1] if len(sys.argv) > 1 else 'wavec'
result = subprocess.run([compiler, '--version'], check=True, text=True, capture_output=True, timeout=10)
match = re.search(r'^wavec\s+(\S+)', result.stdout, re.MULTILINE)
if not match or match[1] != expected:
    raise SystemExit(f'This repository requires wavec {expected}. Use its bundled standard library.')
print(f'wavec {expected}: OK')
