import pytest

def pytest_addoption(parser):
    parser.addoption(
        "--visual-delay",
        action="store_true",
        default=False,
        help="Activa una pausa de 1 segundo entre tests para visualizar cambios en el frontend"
    )

@pytest.fixture
def visual_delay(request):
    return request.config.getoption("--visual-delay")
