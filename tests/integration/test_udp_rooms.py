import socket
import struct
import time
import pytest
import os

UDP_IP = "127.0.0.1"
UDP_PORT = 1884
MAGIC_BYTE = 0x5A

# Mapeo de estados del protocolo: 1 = Abierto (DoorOpen), 0 = Cerrado (DoorClosed)
DOOR_OPEN = 1
DOOR_CLOSED = 0


def build_sensor_packet(room_id: str, state: int) -> bytes:
    """
    Construye la trama de telemetría binaria esperada por RoomsUDPController:
    - [0]   : Magic Byte (0x5A)
    - [1:4] : Room ID en ASCII (exactamente 3 caracteres)
    - [4]   : Estado de puerta (0 = Cerrado, 1 = Abierto)
    """
    room_bytes = room_id.encode("ascii")
    if len(room_bytes) != 3:
        raise ValueError(f"Room ID debe tener 3 caracteres, recibido: {room_id}")
    return struct.pack("!B3sB", MAGIC_BYTE, room_bytes, state)


@pytest.fixture
def udp_client():
    """Socket UDP con timeout de recepción de 2 segundos."""
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.settimeout(2.0)
    yield sock
    sock.close()


@pytest.mark.parametrize(
    "room_id, door_state, state_label, should_exist",
    [
        ("LPA", DOOR_OPEN, "ABIERTO", True),
        ("LPA", DOOR_CLOSED, "CERRADO", True),
        ("OFI", DOOR_OPEN, "ABIERTO", True),
        ("OFI", DOOR_CLOSED, "CERRADO", True),
        ("NUL", DOOR_OPEN, "ABIERTO", False),
        ("NUL", DOOR_CLOSED, "CERRADO", False),
    ],
)
def test_send_room_udp_telemetry(
    udp_client, room_id: str, door_state: int, state_label: str, should_exist: bool
):
    # Envío del datagrama UDP
    packet = build_sensor_packet(room_id, door_state)
    udp_client.sendto(packet, (UDP_IP, UDP_PORT))
    if not should_exist:
        with pytest.raises(socket.timeout):
            udp_client.recvfrom(16)
        return

    # Recepción del ACK emitido por el servidor
    data, _ = udp_client.recvfrom(16)

    # Validaciones del protocolo:
    # 1. El paquete de respuesta mide al menos 2 bytes [0x5A, policy]
    assert len(data) >= 2, f"Respuesta UDP demasiado corta: {data.hex()}"
    
    # 2. Magic byte de respuesta coincide con 0x5A
    assert data[0] == MAGIC_BYTE, f"Magic byte inválido: {hex(data[0])}"
    
    # 3. Política energética válida (1=DEFAULT, 2=ENERGY_SAVER, 3=ULTRA_ENERGY_SAVER)
    policy = data[1]
    assert policy in (1, 2, 3), f"Modo de política desconocido recibido: {policy}"

    # Pausa de 1 segundo entre ejecuciones para permitir la visualización y broadcast en el frontend
    if os.getenv("TEST_VISUAL_DELAY"):
        time.sleep(1.0)