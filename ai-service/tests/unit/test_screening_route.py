from unittest.mock import AsyncMock

from fastapi.testclient import TestClient

from domain.entities import ScreeningResult
from infrastructure.config import Settings
from interfaces.http import create_app


def _client(use_case: AsyncMock) -> TestClient:
    app = create_app(Settings(_env_file=None))  # type: ignore[call-arg]
    app.state.screening = use_case
    return TestClient(app)


def test_check_clean_text_returns_ok_true() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(ok=True, reason=None, categories={"profanity": 0.05})
    )
    client = _client(use_case)
    resp = client.post("/v1/screening/check", json={"text": "hi", "kind": "post"})
    assert resp.status_code == 200
    body = resp.json()
    assert body["ok"] is True
    assert body["reason"] is None
    assert body["categories"]["profanity"] == 0.05


def test_check_blocked_returns_reason() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(
            ok=False,
            reason="profanity",
            categories={"profanity": 0.9},
        )
    )
    client = _client(use_case)
    resp = client.post("/v1/screening/check", json={"text": "bad", "kind": "answer"})
    assert resp.status_code == 200
    assert resp.json()["reason"] == "profanity"


def test_check_unavailable_propagates_reason() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ScreeningResult(ok=False, reason="screening_unavailable", categories={})
    )
    client = _client(use_case)
    resp = client.post("/v1/screening/check", json={"text": "x", "kind": "question"})
    assert resp.status_code == 200
    assert resp.json()["reason"] == "screening_unavailable"


def test_check_validates_kind() -> None:
    use_case = AsyncMock()
    client = _client(use_case)
    resp = client.post("/v1/screening/check", json={"text": "x", "kind": "bogus"})
    assert resp.status_code == 422
