#ifndef WIFI_MANAGER_H
#define WIFI_MANAGER_H

#include <Arduino.h>
#include <IPAddress.h>

enum class WifiStatus : uint8_t
{
    Idle,
    Connecting,
    Connected,
    ConnectionFailed
};

struct WifiSessionInfo
{
    IPAddress ip;
    IPAddress gateway;
    IPAddress subnet;
    IPAddress dns;
    uint8_t bssid[6];
    uint8_t channel;
};

/**
 * @brief Establishes a link with an opportunistic fast-path strategy and guaranteed fallback.
 * @param ssid Target service set identifier.
 * @param password Network access passphrase or user account password.
 * @param username Optional username for WPA2-Enterprise (PEAP/MSCHAPv2). Pass nullptr for WPA2-PSK.
 * @param identity Optional anonymous outer identity for PEAP phase 1. Pass nullptr to inherit username.
 * @param fastPathConfig Optional pointer to cached session parameters.
 * @param fastTimeoutMs Maximum allowable connection window for fast channel/BSSID lock.
 * @param fallbackTimeoutMs Maximum allowable connection window for full broadcast negotiation.
 * @return True if station successfully acquired network access.
 */
bool wifi_connect_resilient(
    const char *ssid,
    const char *password,
    const char *username = nullptr,
    const char *identity = nullptr,
    const WifiSessionInfo *fastPathConfig = nullptr,
    uint32_t fastTimeoutMs = 1200,
    uint32_t fallbackTimeoutMs = 4500);

/**
 * @brief Extracts active network parameters from the connected interface.
 * @param[out] outSession Destination buffer for active link descriptors.
 * @return True if session information was successfully resolved.
 */
bool wifi_get_session_info(WifiSessionInfo &outSession);

#endif // WIFI_MANAGER_H