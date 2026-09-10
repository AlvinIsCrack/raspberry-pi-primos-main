#ifndef BOOT_H
#define BOOT_H

#include <Arduino.h>

/**
 * @brief Discrete operational phases and diagnostic fault conditions.
 */
enum class BootStatus : uint8_t
{
    InitHardware,
    ConnectingWifi,
    DispatchingUdp,
    Success,
    ErrorWifi,
    ErrorSensor,
    ErrorUdp
};

/**
 * @brief Configures output pin driving the diagnostic LED indicator.
 * @param pin Target GPIO pin identifier.
 * @param activeLow Hardware sink circuit inversion flag.
 */
void boot_init(uint8_t pin = LED_BUILTIN, bool activeLow = true);

/**
 * @brief Transitions the diagnostic engine to a target status code.
 * @param status Target status code.
 */
void boot_set_status(BootStatus status);

/**
 * @brief Emits a finite sequence of diagnostic LED pulses in a blocking manner.
 * @param status Diagnostic operational or fault status to signal.
 * @param repeatCount Total iterations the pulse cadence is presented.
 */
void boot_signal_blocking(BootStatus status, uint8_t repeatCount = 1);

#endif // BOOT_H