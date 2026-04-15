from __future__ import annotations

import os
import sys

GEN_DIR = os.path.join(os.path.dirname(__file__), "gen")
if GEN_DIR not in sys.path:
    sys.path.insert(0, GEN_DIR)

