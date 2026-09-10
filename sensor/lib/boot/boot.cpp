#include "boot.h"

enum class IndicatorMode : uint8_t
{
    Off,
    PulseBurst
};

struct SequencePattern
{
    IndicatorMode mode;
    uint8_t pulseCount;
    uint16_t onDurationMs;
    uint16_t offDurationMs;
    uint16_t pauseDurationMs;
    const char *logMessage;
};

enum class BurstSubState : uint8_t
{
    PulseHigh,
    PulseLow,
    SequencePause
};

static uint8_t _ledPin = LED_BUILTIN;
static bool _activeLow = true;
static BootStatus _currentStatus = BootStatus::InitHardware;

static uint32_t _lastTransitionTime = 0;
static uint8_t _currentPulse = 0;
static BurstSubState _burstState = BurstSubState::PulseHigh;

static inline void writePinState(bool stateHigh)
{
    digitalWrite(_ledPin, (_activeLow ? !stateHigh : stateHigh) ? HIGH : LOW);
}

static const SequencePattern *resolvePattern(BootStatus status)
{
    // Mode, PulseCount, OnDurationMs, OffDurationMs, PauseDurationMs, LogMessage
    static constexpr SequencePattern patterns[] = {
        // Ephemeral progress stages (Fast cadences: 100ms on/off, 500ms pause)
        {IndicatorMode::PulseBurst, 1, 100, 100, 500, "[BOOT] Initializing hardware core..."},
        {IndicatorMode::PulseBurst, 2, 100, 100, 500, "[BOOT] Negotiating WiFi link..."},
        {IndicatorMode::PulseBurst, 3, 100, 100, 500, "[BOOT] Synchronizing UDP session..."},
        // Operational idle state (LED suppressed to conserve power)
        {IndicatorMode::Off, 0, 0, 0, 0, "[BOOT] Boot sequence completed successfully. Systems operational."},
        // Terminal fault conditions (Deliberate cadences: 200ms on/off, 1500ms pause)
        {IndicatorMode::PulseBurst, 2, 200, 200, 1500, "[BOOT] Critical error (Code 2): WiFi link failure."},
        {IndicatorMode::PulseBurst, 3, 200, 200, 1500, "[BOOT] Peripheral error (Code 3): Sensor communication failure."},
        {IndicatorMode::PulseBurst, 4, 200, 200, 1500, "[BOOT] Critical error (Code 4): UDP server unreachable after 3 attempts."}};

    return &patterns[static_cast<size_t>(status)];
}

void boot_init(uint8_t pin, bool activeLow)
{
    _ledPin = pin;
    _activeLow = activeLow;
    pinMode(_ledPin, OUTPUT);
    boot_set_status(BootStatus::InitHardware);
}

void boot_set_status(BootStatus status)
{
    _currentStatus = status;
    _lastTransitionTime = millis();
    _currentPulse = 0;

    const SequencePattern *pattern = resolvePattern(status);

    if (pattern->mode == IndicatorMode::PulseBurst)
    {
        _burstState = BurstSubState::PulseHigh;
        writePinState(true);
    }
    else
    {
        writePinState(false);
    }

    if (pattern->logMessage != nullptr)
    {
        Serial.println(pattern->logMessage);
    }
}

void boot_signal_blocking(BootStatus status, uint8_t repeatCount)
{
    const SequencePattern *pattern = resolvePattern(status);
    if (pattern->mode == IndicatorMode::Off)
    {
        writePinState(false);
        return;
    }

    for (uint8_t i = 0; i < repeatCount; ++i)
    {
        for (uint8_t pulse = 0; pulse < pattern->pulseCount; ++pulse)
        {
            writePinState(true);
            delay(pattern->onDurationMs);
            writePinState(false);
            delay(pattern->offDurationMs);
        }
        delay(pattern->pauseDurationMs);
    }
}