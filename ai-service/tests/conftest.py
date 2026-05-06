import pytest

from infrastructure.config import Settings
from infrastructure.observability import configure_logging


@pytest.fixture
def settings() -> Settings:
    return Settings(
        _env_file=None,  # type: ignore[call-arg]
    )


@pytest.fixture(autouse=True, scope="session")
def _logging() -> None:
    configure_logging(Settings(_env_file=None).logging)  # type: ignore[call-arg]
