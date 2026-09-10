#include "bootstrap.h"
#include <boot.h>
#include <wifi_manager.h>
#include "config.h"
#include <ESP8266WiFi.h>
#include "core/udp/udp_client.h"

bool bootstrap_run_sequence(bool isColdBoot, const NetworkCache &cache)
{
    boot_init(Config::Hardware::StatusLedPin, Config::Hardware::StatusLedActiveLow);
    boot_set_status(BootStatus::InitHardware);

    if (isColdBoot)
    {
        boot_signal_blocking(BootStatus::InitHardware, 1);
    }

    boot_set_status(BootStatus::ConnectingWifi);

    WifiSessionInfo fastConfig;
    WifiSessionInfo *fastConfigPtr = nullptr;

    if (cache.isValid)
    {
        fastConfig.ip = IPAddress(cache.ip);
        fastConfig.gateway = IPAddress(cache.gateway);
        fastConfig.subnet = IPAddress(cache.subnet);
        fastConfig.dns = IPAddress(cache.dns);
        fastConfig.channel = cache.channel;
        memcpy(fastConfig.bssid, cache.bssid, sizeof(fastConfig.bssid));
        fastConfigPtr = &fastConfig;
    }

    if (!wifi_connect_resilient(
            Config::Wifi::Ssid,
            Config::Wifi::Password,
            fastConfigPtr,
            Config::Wifi::FastConnectTimeoutMs,
            Config::Wifi::FallbackTimeoutMs))
    {
        boot_set_status(BootStatus::ErrorWifi);
        boot_signal_blocking(BootStatus::ErrorWifi, 3);
        return false;
    }

    // Capture and persist active link descriptors immediately upon link resolution
    WifiSessionInfo currentSession;
    if (wifi_get_session_info(currentSession))
    {
        NetworkCache activeCache;
        activeCache.ip = static_cast<uint32_t>(currentSession.ip);
        activeCache.gateway = static_cast<uint32_t>(currentSession.gateway);
        activeCache.subnet = static_cast<uint32_t>(currentSession.subnet);
        activeCache.dns = static_cast<uint32_t>(currentSession.dns);
        activeCache.channel = currentSession.channel;
        memcpy(activeCache.bssid, currentSession.bssid, sizeof(activeCache.bssid));
        activeCache.isValid = true;

        state_tracker_update_network_cache(activeCache);
    }

    udp_init(Config::Udp::LocalPort);
    boot_set_status(BootStatus::DispatchingUdp);
    return true;
}