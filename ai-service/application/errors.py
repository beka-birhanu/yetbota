from __future__ import annotations


class AppError(Exception):
    pass


class InvalidQuery(AppError):
    pass


class Misconfigured(AppError):
    pass


class UpstreamUnavailable(AppError):
    pass

