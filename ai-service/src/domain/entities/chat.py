from pydantic import BaseModel, ConfigDict, Field

from domain.entities.citation import Citation


class ChatQuery(BaseModel):
    model_config = ConfigDict(frozen=True)

    text: str
    user_id: str | None = None


class ChatResponse(BaseModel):
    model_config = ConfigDict(frozen=True)

    answer: str
    citations: list[Citation] = Field(default_factory=list)
