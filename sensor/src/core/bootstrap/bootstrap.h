#ifndef BOOTSTRAP_H
#define BOOTSTRAP_H

#include <Arduino.h>
#include "core/state/state_tracker.h"

/**
 * @brief Coordinates diagnostic signalling, Wi-Fi link acquisition, and UDP socket initialization.
 * @param isColdBoot Enables initial visual hardware diagnostic cadence.
 * @param cache Cached link parameters used for fast-path association.
 * @return True if Wi-Fi associated and the UDP socket is open.
 */
bool bootstrap_run_sequence(bool isColdBoot, const NetworkCache &cache);

#endif // BOOTSTRAP_H