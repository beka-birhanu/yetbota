from __future__ import annotations

import os
import sys

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", "..", ".."))
COMMON_PY_GEN = os.path.join(REPO_ROOT, "common", "proto", "generated", "python")
if COMMON_PY_GEN not in sys.path:
    sys.path.insert(0, COMMON_PY_GEN)

