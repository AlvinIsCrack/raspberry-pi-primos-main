#include "wifi_manager.h"
#include <ESP8266WiFi.h>

extern "C"
{
#include "user_interface.h"
#include "wpa2_enterprise.h"
#include "c_types.h"
}

static constexpr uint32_t FastConnectBudgetMs = 2500;

/**
 * @brief Configures the ESP8266 station peripheral for WPA2-Enterprise PEAP authentication.
 * @param username User credential for EAP-MSCHAPv2 authentication.
 * @param password Password associated with user credential.
 * @param identity Outer anonymous identity for phase 1 negotiation.
 * @return True if all native SDK calls succeeded.
 */
static bool configure_wpa2_enterprise(const char *username, const char *password, const char *identity)
{
    wifi_station_disconnect();

    // Invalidate stale security contexts before injecting new parameters
    wifi_station_clear_cert_key();
    wifi_station_clear_enterprise_ca_cert();
    wifi_station_clear_enterprise_identity();
    wifi_station_clear_enterprise_username();
    wifi_station_clear_enterprise_password();
    wifi_station_clear_enterprise_new_password();

    if (wifi_station_set_wpa2_enterprise_auth(1) != 0)
    {
        return false;
    }

    const char *outerIdentity = (identity != nullptr && identity[0] != '\0') ? identity : username;

    if (wifi_station_set_enterprise_identity(reinterpret_cast<uint8_t *>(const_cast<char *>(outerIdentity)), strlen(outerIdentity)) != 0)
    {
        return false;
    }

    if (wifi_station_set_enterprise_username(reinterpret_cast<uint8_t *>(const_cast<char *>(username)), strlen(username)) != 0)
    {
        return false;
    }

    if (wifi_station_set_enterprise_password(reinterpret_cast<uint8_t *>(const_cast<char *>(password)), strlen(password)) != 0)
    {
        return false;
    }

    return true;
}

static bool await_association(uint32_t timeoutBudgetMs)
{
    const uint32_t startMs = millis();
    while ((WiFi.status() != WL_CONNECTED || WiFi.localIP() == IPAddress(0, 0, 0, 0) || WiFi.localIP() == INADDR_NONE) && (millis() - startMs < timeoutBudgetMs))
    {
        delay(10);
    }
    return (WiFi.status() == WL_CONNECTED) && (WiFi.localIP() != IPAddress(0, 0, 0, 0)) && (WiFi.localIP() != INADDR_NONE);
}

bool wifi_connect_resilient(
    const char *ssid,
    const char *password,
    const char *username,
    const char *identity,
    const WifiSessionInfo *fastPathConfig,
    uint32_t fastTimeoutMs,
    uint32_t fallbackTimeoutMs)
{
    WiFi.forceSleepWake();
    delay(10);
    WiFi.persistent(false);
    WiFi.mode(WIFI_STA);

    const bool isEnterprise = (username != nullptr && username[0] != '\0');

    if (isEnterprise)
    {
        Serial.printf("[WIFI] Configuring WPA2-Enterprise context for SSID: %s\n", ssid);
        if (!configure_wpa2_enterprise(username, password, identity))
        {
            Serial.println("[WIFI] Failed to set up 802.1X credentials.");
            return false;
        }
    }

    // Opportunistic fast path execution
    if (fastPathConfig != nullptr && fastPathConfig->channel > 0)
    {
        Serial.printf("[WIFI] Attempting fast link: Ch %u\n", fastPathConfig->channel);
        WiFi.config(
            fastPathConfig->ip,
            fastPathConfig->gateway,
            fastPathConfig->subnet,
            fastPathConfig->dns);

        if (isEnterprise)
        {
            // Under WPA2-Enterprise, station begins link solely via SSID parameter
            WiFi.begin(ssid, nullptr, fastPathConfig->channel, fastPathConfig->bssid, true);
        }
        else
        {
            WiFi.begin(ssid, password, fastPathConfig->channel, fastPathConfig->bssid, true);
        }

        if (await_association(fastTimeoutMs))
        {
            Serial.printf("[WIFI] Fast link established in %lu ms. IP: %s\n",
                          millis(),
                          WiFi.localIP().toString().c_str());
            return true;
        }

        Serial.println("[WIFI] Fast path acquisition timed out. Reverting to full scan fallback.");
        WiFi.disconnect(false);
        delay(20);

        if (isEnterprise)
        {
            // Re-arm Enterprise driver state post-disconnect
            configure_wpa2_enterprise(username, password, identity);
        }
    }

    // Full scan and standard DHCP resolution fallback
    Serial.println("[WIFI] Initiating full broadcast scan with standard DHCP negotiation...");
    WiFi.config(0U, 0U, 0U);

    if (isEnterprise)
    {
        WiFi.begin(ssid);
    }
    else
    {
        WiFi.begin(ssid, password);
    }

    if (await_association(fallbackTimeoutMs))
    {
        Serial.printf("[WIFI] Fallback link synchronized successfully. Assigned IP: %s\n",
                      WiFi.localIP().toString().c_str());
        return true;
    }

    Serial.printf("[WIFI] Connection aborted: Timeout bound of %u ms exceeded.\n", fallbackTimeoutMs);
    WiFi.disconnect(true);
    if (isEnterprise)
    {
        wifi_station_set_wpa2_enterprise_auth(0);
    }
    return false;
}

bool wifi_get_session_info(WifiSessionInfo &outSession)
{
    if (WiFi.status() != WL_CONNECTED)
    {
        return false;
    }

    outSession.ip = WiFi.localIP();
    outSession.gateway = WiFi.gatewayIP();
    outSession.subnet = WiFi.subnetMask();
    outSession.dns = WiFi.dnsIP();
    outSession.channel = WiFi.channel();

    const uint8_t *activeBssid = WiFi.BSSID();
    if (activeBssid != nullptr)
    {
        memcpy(outSession.bssid, activeBssid, 6);
    }
    else
    {
        memset(outSession.bssid, 0, 6);
    }

    return true;
}