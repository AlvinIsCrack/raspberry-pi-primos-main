#ifndef CONFIG_H
#define CONFIG_H

#include <Arduino.h>
#include <IPAddress.h>

namespace Config
{
    namespace Wifi
    {
        inline constexpr const char *Ssid = "fh_fa38f8_";
        inline constexpr const char *Password = "wlan05c707";

        /** Set true to enable WPA2 802.1X PEAP/MSCHAPv2 authentication (e.g., eduroam). */
        inline constexpr bool EnterpriseEnabled = false;
        /** Identity / Username for 802.1X authentication; leave nullptr for standard WPA2-PSK. */
        inline constexpr const char *EnterpriseUsername = nullptr;
        /** Anonymous outer identity for PEAP tunneling; leave nullptr or empty to reuse username. */
        inline constexpr const char *EnterpriseIdentity = nullptr;

        inline constexpr uint32_t FastConnectTimeoutMs = 2000;
        inline constexpr uint32_t FallbackTimeoutMs = 5000;
    }

    namespace Udp
    {
        inline const IPAddress ServerIp(192, 168, 1, 2);
        inline constexpr uint16_t ServerPort = 1884;
        inline constexpr uint16_t LocalPort = 4210;
        inline constexpr const char DoorId[3] = {'L', 'D', 'S'};
        inline constexpr uint8_t MagicByte = 0x5A;
        inline constexpr uint8_t MaxRetries = 3;
        inline constexpr uint32_t RetryTimeoutMs = 100;
    }

    namespace Sensor
    {
        inline constexpr uint8_t LatchSwitchPin = D5;
        inline constexpr bool ActiveLow = true;
        inline constexpr bool UseInternalPullup = true;
    }

    namespace Hardware
    {
        inline constexpr uint8_t StatusLedPin = LED_BUILTIN;
        inline constexpr bool StatusLedActiveLow = true;
        inline constexpr uint32_t SerialBaudRate = 115200;

        /** Polling cycle cadence in seconds between physical pin evaluations. */
        inline constexpr uint32_t DefaultSleepSec = 15;
        /** Enforced transmission window for broker heartbeat verification. */
        inline constexpr uint32_t DefaultHeartbeatSec = 15 * 60;

        inline void resolvePolicy(uint8_t mode, uint32_t &outSleepSec, uint32_t &outHeartbeatCycles)
        {
            uint32_t hbSec = DefaultHeartbeatSec;
            switch (mode)
            {
            case 2: // Energy Saver
                outSleepSec = 20;
                hbSec = 30 * 60;
                break;
            case 3: // Ultra Energy Saver
                outSleepSec = 30;
                hbSec = 60 * 60;
                break;
            case 1: // Default
            default:
                outSleepSec = DefaultSleepSec;
                hbSec = DefaultHeartbeatSec;
                break;
            }
            outHeartbeatCycles = hbSec / outSleepSec;
        }
    }
}

#endif // CONFIG_H