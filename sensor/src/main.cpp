#include <Arduino.h>
#include <ESP8266WiFi.h>
#include <boot.h>

#include "config.h"
#include "core/bootstrap/bootstrap.h"
#include "core/state/state_tracker.h"
#include "core/udp/udp_client.h"

#define DEBUG 1
#if !DEBUG
#define Serial \
  if (false)   \
  Serial
#endif

void setup()
{
  Serial.begin(Config::Hardware::SerialBaudRate);
  const DispatchDecision decision = state_tracker_evaluate();

  uint32_t sleepSec = Config::Hardware::DefaultSleepSec;
  uint32_t hbCycles = 60;
  Config::Hardware::resolvePolicy(decision.currentPolicy, sleepSec, hbCycles);

  if (!decision.shouldDispatch)
  {
    state_tracker_commit_skip();
    ESP.deepSleep(static_cast<uint64_t>(sleepSec) * 1000000ULL);
    return;
  }

  if (bootstrap_run_sequence(decision.isColdBoot, decision.networkCache))
  {
    uint8_t receivedPolicy = 0;
    const bool txSuccess = udp_dispatch_telemetry(decision.isDoorLocked, receivedPolicy);
    if (txSuccess)
    {
      if (receivedPolicy >= 1 && receivedPolicy <= 3)
      {
        Config::Hardware::resolvePolicy(receivedPolicy, sleepSec, hbCycles);
      }
      boot_set_status(BootStatus::Success);
      state_tracker_commit_dispatch(decision.isDoorLocked, receivedPolicy);
    }
    else
    {
      boot_set_status(BootStatus::ErrorUdp);
      boot_signal_blocking(BootStatus::ErrorUdp, 3);
      state_tracker_commit_skip();
    }
    udp_stop();
    WiFi.disconnect(true);
    delay(5);
  }
  else
  {
    WiFi.disconnect(true);
    delay(1);
  }

  Serial.printf("[PWR] Deep sleep duration: %u seconds...\n", sleepSec);
  ESP.deepSleep(static_cast<uint64_t>(sleepSec) * 1000000ULL);
}

void loop()
{
}