from pydantic import BaseModel, ConfigDict

class IncomingMessage(BaseModel):
    model_config = ConfigDict(frozen=True, arbitrary_types_allowed=True)

    body: bytes
    delivery_count: int
