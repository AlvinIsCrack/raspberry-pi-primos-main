#include "udp_client.h"
#include "config.h"
#include <WiFiUdp.h>

static WiFiUDP _udp;

bool udp_init(uint16_t localPort)
{
    return _udp.begin(localPort) == 1;
}

bool udp_dispatch_telemetry(bool isDoorLocked, uint8_t &outPolicy)
{
    // Binary packet payload (5 bytes): [0x5A, 'L', 'D', 'S', state]
    uint8_t packet[5];
    packet[0] = Config::Udp::MagicByte;
    packet[1] = Config::Udp::DoorId[0];
    packet[2] = Config::Udp::DoorId[1];
    packet[3] = Config::Udp::DoorId[2];
    packet[4] = isDoorLocked ? 0 : 1; // 0 = Locked/Closed, 1 = Open

    uint8_t replyBuf[16];

    for (uint8_t attempt = 1; attempt <= Config::Udp::MaxRetries; ++attempt)
    {
        // Flush any stale inbound datagrams queued in the socket buffer
        while (_udp.parsePacket() > 0)
        {
            _udp.flush();
        }

        _udp.beginPacket(Config::Udp::ServerIp, Config::Udp::ServerPort);
        _udp.write(packet, sizeof(packet));
        _udp.endPacket();

        const uint32_t startWait = millis();
        while (millis() - startWait < Config::Udp::RetryTimeoutMs)
        {
            // Allow the network stack to service low-level packet reception
            yield();

            const int packetSize = _udp.parsePacket();
            if (packetSize >= 2)
            {
                const int bytesRead = _udp.read(replyBuf, sizeof(replyBuf));
                if (bytesRead >= 2 && replyBuf[0] == Config::Udp::MagicByte)
                {
                    outPolicy = replyBuf[1];
                    Serial.printf("[UDP] Inbound ACK on attempt %u. Policy: %u\n", attempt, outPolicy);
                    return true;
                }
            }
            delay(5);
        }
        Serial.printf("[UDP] Attempt %u exhausted without ACK.\n", attempt);
    }

    return false;
}

void udp_stop()
{
    _udp.stop();
}