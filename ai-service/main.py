from __future__ import annotations

import argparse
import os
import sys

import uvicorn

from infra.config import get_settings
from interfaces.grpc.server import serve as serve_grpc
from interfaces.rest.app import create_app


def _parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(prog="ai-service")
    sub = parser.add_subparsers(dest="cmd", required=True)

    rest = sub.add_parser("rest")
    rest.add_argument("--host", default=None)
    rest.add_argument("--port", type=int, default=None)
    rest.add_argument("--reload", action="store_true")

    sub.add_parser("grpc")
    return parser.parse_args(argv)


def _ensure_project_root_on_syspath() -> None:
    root = os.path.dirname(__file__)
    if root not in sys.path:
        sys.path.insert(0, root)


def main(argv: list[str] | None = None) -> None:
    _ensure_project_root_on_syspath()
    args = _parse_args(sys.argv[1:] if argv is None else argv)
    settings = get_settings()

    if args.cmd == "rest":
        host = args.host or settings.rest_host
        port = args.port or settings.rest_port
        uvicorn.run(create_app(), host=host, port=port, reload=bool(args.reload))
        return

    if args.cmd == "grpc":
        serve_grpc()
        return

    raise SystemExit(2)


if __name__ == "__main__":
    main()

