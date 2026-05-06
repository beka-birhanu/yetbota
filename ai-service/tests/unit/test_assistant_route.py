from unittest.mock import AsyncMock

from fastapi.testclient import TestClient

from application.errors import LLMUnavailable
from domain.entities import ChatResponse, Citation
from infrastructure.config import Settings
from interfaces.http import create_app


def _client(use_case: AsyncMock) -> TestClient:
    app = create_app(Settings(_env_file=None))  # type: ignore[call-arg]
    app.state.assistant = use_case
    return TestClient(app)


def test_chat_returns_answer_and_citations() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(
        return_value=ChatResponse(
            answer="The answer is X.",
            citations=[
                Citation(source_id="p1", kind="post", text="hello", score=0.9),
            ],
        )
    )
    client = _client(use_case)
    resp = client.post("/v1/assistant/chat", json={"text": "where?", "user_id": None})
    assert resp.status_code == 200
    body = resp.json()
    assert body["answer"] == "The answer is X."
    assert body["citations"][0]["source_id"] == "p1"
    assert body["citations"][0]["kind"] == "post"


def test_chat_returns_503_on_llm_unavailable() -> None:
    use_case = AsyncMock()
    use_case.execute = AsyncMock(side_effect=LLMUnavailable("down"))
    client = _client(use_case)
    resp = client.post("/v1/assistant/chat", json={"text": "x"})
    assert resp.status_code == 503
    assert "currently unavailable" in resp.json()["detail"]


def test_chat_validates_request_body() -> None:
    use_case = AsyncMock()
    client = _client(use_case)
    resp = client.post("/v1/assistant/chat", json={})
    assert resp.status_code == 422
