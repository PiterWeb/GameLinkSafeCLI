package proxy

import "github.com/pion/webrtc/v3"

// SendThroughUDP sends data through a UDP connection (used for host logic)
func SendThroughUDP(port uint, entryChannel <-chan []byte, exitChannel *webrtc.DataChannel) error {
	return sendThroughHost(UDP, port, entryChannel, exitChannel)
}

// ServeThroughUDP serves as a UDP server (used for client logic)
func ServeThroughUDP(port uint, entryChannel <-chan []byte, exitChannel *webrtc.DataChannel) error {
	return serveThroughClient(UDP, port, entryChannel, exitChannel)
}
