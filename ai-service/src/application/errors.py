from __future__ import annotations


class AIServiceError(Exception):
    code: str = "UNKNOWN_ERROR"
    transient: bool = False

    def __init__(self, message: str = "", *, cause: Exception | None = None) -> None:
        super().__init__(message)
        self.message = message
        self.__cause__ = cause


class ConfigError(AIServiceError):
    code = "CONFIG_ERROR"


class MessageMalformed(AIServiceError):
    code = "MESSAGE_MALFORMED"


class MetadataError(AIServiceError):
    code = "METADATA_ERROR"


class EmbeddingFailed(AIServiceError):
    code = "EMBED_FAILED"
    transient = True


class SimilaritySearchFailed(AIServiceError):
    code = "SIMILARITY_FAILED"
    transient = True


class IndexingFailed(AIServiceError):
    code = "INDEXING_GAP"


class LLMUnavailable(AIServiceError):
    code = "LLM_UNAVAILABLE"
    transient = True


class ScreeningUnavailable(AIServiceError):
    code = "SCREENING_UNAVAILABLE"
    transient = True
