package proxy

import (
	"github.com/pion/webrtc/v3"
)

// SendThroughUDP sends data through a UDP connection (used for host logic)
func SendThroughUDP(port uint, proxyChan <-chan []byte, exitChannel *webrtc.DataChannel) error {
	return sendThroughHostUDP(port, proxyChan, exitChannel)
}

// ServeThroughUDP serves as a UDP server (used for client logic)
func ServeThroughUDP(port uint, proxyChan <-chan []byte, exitChannel *webrtc.DataChannel) error {
	return serveThroughClientUDP(port, proxyChan, exitChannel)
}
