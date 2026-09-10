#ifndef STATE_TRACKER_H
#define STATE_TRACKER_H

#include <Arduino.h>
#include <IPAddress.h>

/**
 * @brief Cached physical and network-layer parameters for rapid reconnection.
 */
struct NetworkCache
{
    uint32_t ip;
    uint32_t gateway;
    uint32_t subnet;
    uint32_t dns;
    uint8_t bssid[6];
    uint8_t channel;
    bool isValid;
};

/**
 * @brief Dispatch evaluation result containing operational context.
 */
struct DispatchDecision
{
    bool shouldDispatch;
    bool isHeartbeat;
    bool isDoorLocked;
    bool isColdBoot;
    NetworkCache networkCache;
    uint8_t currentPolicy;
};

/**
 * @brief Initializes sensor peripherals and evaluates dispatch criteria against RTC state.
 * @return Struct detailing whether network wake-up is required and current door latch state.
 */
DispatchDecision state_tracker_evaluate();

/**
 * @brief Persists dispatch success, updates tracking state, and applies negotiated policy atomically.
 * @param doorLocked Current door latch logical state.
 * @param policyMode Optional policy mode received via server ACK (1-3), or 0 to retain current policy.
 */
void state_tracker_commit_dispatch(bool doorLocked, uint8_t policyMode = 0);

/**
 * @brief Increments heartbeat cycle counter and preserves state when sleep is immediately resumed.
 */
void state_tracker_commit_skip();

/**
 * @brief Commits updated radio and network descriptors into RTC memory independently.
 * @param cache Verified link parameters to persist.
 */
void state_tracker_update_network_cache(const NetworkCache &cache);

/**
 * @brief Manually overrides the current operational power policy in RTC memory.
 * @param policyMode Target policy mode identifier (1 to 3).
 */
void state_tracker_set_policy(uint8_t policyMode);

/**
 * @brief Retrieves the active operational power policy mode.
 * @return Active policy mode identifier (1 to 3).
 */
uint8_t state_tracker_get_policy();

#endif // STATE_TRACKER_H