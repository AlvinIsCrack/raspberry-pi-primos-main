#ifndef UDP_CLIENT_H
#define UDP_CLIENT_H

#include <Arduino.h>
#include <IPAddress.h>

/**
 * @brief Binds and initializes the local UDP socket.
 * @param localPort Inbound port for datagram reception.
 * @return True if socket binding succeeded.
 */
bool udp_init(uint16_t localPort);

/**
 * @brief Dispatches the binary sensor payload and awaits an ACK with policy directives.
 * @param isDoorLocked Physical state of the monitored latch switch.
 * @param[out] outPolicy Receives remote policy index (1, 2, or 3) if confirmed by server.
 * @return True if a valid ACK was parsed within the retransmission window.
 */
bool udp_dispatch_telemetry(bool isDoorLocked, uint8_t &outPolicy);

/**
 * @brief Closes and unbinds the local UDP socket.
 */
void udp_stop();

#endif // UDP_CLIENT_H