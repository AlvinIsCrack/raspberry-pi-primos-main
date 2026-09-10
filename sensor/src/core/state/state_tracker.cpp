#include "state_tracker.h"
#include "config.h"

namespace
{
    inline constexpr uint32_t RtcMagicKey = 0x5A5AA5A6;
    inline constexpr uint32_t RtcUserMemoryOffset = 64;

    struct alignas(4) RtcSnapshot
    {
        uint32_t magic;
        uint32_t heartbeatCycles;
        uint8_t lastDoorState;
        uint8_t reserved[3];
        uint8_t currentPolicy;

        // Network session cache
        uint32_t ip;
        uint32_t gateway;
        uint32_t subnet;
        uint32_t dns;
        uint8_t bssid[6];
        uint8_t channel;
        uint8_t cacheValid;
    };

    static RtcSnapshot _snapshot;

    bool read_physical_sensor()
    {
        const int rawValue = digitalRead(Config::Sensor::LatchSwitchPin);
        return Config::Sensor::ActiveLow ? (rawValue == LOW) : (rawValue == HIGH);
    }
}

DispatchDecision state_tracker_evaluate()
{
    if (Config::Sensor::UseInternalPullup)
    {
        pinMode(Config::Sensor::LatchSwitchPin, INPUT_PULLUP);
    }
    else
    {
        pinMode(Config::Sensor::LatchSwitchPin, INPUT);
    }

    const bool currentState = read_physical_sensor();

    ESP.rtcUserMemoryRead(RtcUserMemoryOffset, reinterpret_cast<uint32_t *>(&_snapshot), sizeof(RtcSnapshot));

    const bool isColdBoot = (_snapshot.magic != RtcMagicKey);

    if (isColdBoot)
    {
        memset(&_snapshot, 0, sizeof(RtcSnapshot));
        _snapshot.magic = RtcMagicKey;
        _snapshot.heartbeatCycles = 0;
        // Sentinel value ensuring the first wake-up unconditionally triggers a state delta dispatch
        _snapshot.lastDoorState = 0xFF;
        _snapshot.currentPolicy = 1;
        _snapshot.cacheValid = 0;

        ESP.rtcUserMemoryWrite(RtcUserMemoryOffset, reinterpret_cast<uint32_t *>(&_snapshot), sizeof(RtcSnapshot));

        return {
            true,         // shouldDispatch: unconditionally true on cold boot
            false,        // isHeartbeat
            currentState, // isDoorLocked
            true,         // isColdBoot
            {0, 0, 0, 0, {0}, 0, false},
            1 // currentPolicy default
        };
    }

    NetworkCache cache{
        _snapshot.ip,
        _snapshot.gateway,
        _snapshot.subnet,
        _snapshot.dns,
        {0},
        _snapshot.channel,
        _snapshot.cacheValid == 1};
    memcpy(cache.bssid, _snapshot.bssid, sizeof(cache.bssid));

    const bool previousState = (_snapshot.lastDoorState != 0);
    // Evaluates true if memory holds the sentinel uninitialized state (0xFF) or physical logic toggled
    const bool stateChanged = (_snapshot.lastDoorState == 0xFF) || (currentState != previousState);

    uint32_t sleepSec = Config::Hardware::DefaultSleepSec;
    uint32_t maxCycles = Config::Hardware::DefaultHeartbeatSec / sleepSec;
    Config::Hardware::resolvePolicy(_snapshot.currentPolicy, sleepSec, maxCycles);

    const bool heartbeatExpired = (_snapshot.heartbeatCycles >= maxCycles);

    return {
        stateChanged || heartbeatExpired,
        heartbeatExpired && !stateChanged,
        currentState,
        false,
        cache,
        _snapshot.currentPolicy};
}

void state_tracker_update_network_cache(const NetworkCache &cache)
{
    _snapshot.ip = cache.ip;
    _snapshot.gateway = cache.gateway;
    _snapshot.subnet = cache.subnet;
    _snapshot.dns = cache.dns;
    _snapshot.channel = cache.channel;
    _snapshot.cacheValid = cache.isValid ? 1 : 0;
    memcpy(_snapshot.bssid, cache.bssid, sizeof(_snapshot.bssid));
    ESP.rtcUserMemoryWrite(RtcUserMemoryOffset, reinterpret_cast<uint32_t *>(&_snapshot), sizeof(RtcSnapshot));
}

void state_tracker_commit_dispatch(bool doorLocked, uint8_t policyMode)
{
    _snapshot.lastDoorState = doorLocked ? 1 : 0;
    _snapshot.heartbeatCycles = 0;

    if (policyMode >= 1 && policyMode <= 3)
    {
        _snapshot.currentPolicy = policyMode;
    }

    ESP.rtcUserMemoryWrite(RtcUserMemoryOffset, reinterpret_cast<uint32_t *>(&_snapshot), sizeof(RtcSnapshot));
}

void state_tracker_commit_skip()
{
    _snapshot.heartbeatCycles++;
    ESP.rtcUserMemoryWrite(RtcUserMemoryOffset, reinterpret_cast<uint32_t *>(&_snapshot), sizeof(RtcSnapshot));
}

void state_tracker_set_policy(uint8_t policyMode)
{
    if (policyMode >= 1 && policyMode <= 3)
    {
        _snapshot.currentPolicy = policyMode;
        ESP.rtcUserMemoryWrite(RtcUserMemoryOffset, reinterpret_cast<uint32_t *>(&_snapshot), sizeof(RtcSnapshot));
    }
}

uint8_t state_tracker_get_policy()
{
    return (_snapshot.currentPolicy >= 1 && _snapshot.currentPolicy <= 3) ? _snapshot.currentPolicy : 1;
}