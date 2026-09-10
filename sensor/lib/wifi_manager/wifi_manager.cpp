#include "wifi_manager.h"
#include <ESP8266WiFi.h>

static constexpr uint32_t FastConnectBudgetMs = 2500;

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
    const WifiSessionInfo *fastPathConfig,
    uint32_t fastTimeoutMs,
    uint32_t fallbackTimeoutMs)
{
    WiFi.forceSleepWake();
    delay(10);

    WiFi.persistent(false);
    WiFi.mode(WIFI_STA);

    // Opportunistic fast path execution
    if (fastPathConfig != nullptr && fastPathConfig->channel > 0)
    {
        Serial.printf("[WIFI] Attempting fast link: Ch %u\n", fastPathConfig->channel);
        WiFi.config(
            fastPathConfig->ip,
            fastPathConfig->gateway,
            fastPathConfig->subnet,
            fastPathConfig->dns);
        WiFi.begin(ssid, password, fastPathConfig->channel, fastPathConfig->bssid, true);

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
    }

    // Full scan and standard DHCP resolution fallback
    Serial.println("[WIFI] Initiating full broadcast scan with standard DHCP negotiation...");
    WiFi.config(0U, 0U, 0U);
    WiFi.begin(ssid, password);

    if (await_association(fallbackTimeoutMs))
    {
        Serial.printf("[WIFI] Fallback link synchronized successfully. Assigned IP: %s\n",
                      WiFi.localIP().toString().c_str());
        return true;
    }

    Serial.printf("[WIFI] Connection aborted: Timeout bound of %u ms exceeded.\n", fallbackTimeoutMs);
    WiFi.disconnect(true);
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