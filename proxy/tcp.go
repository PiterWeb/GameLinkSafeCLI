package proxy

import "github.com/pion/webrtc/v3"

// SendThroughTCP sends data through a TCP connection (used for host logic)
func SendThroughTCP(port uint, entryChannel <-chan []byte, exitDataChannel *webrtc.DataChannel) error {
	return sendThroughHost(TCP, port, entryChannel, exitDataChannel)
}

// ServeThroughTCP serves as a TCP server (used for client logic)
func ServeThroughTCP(port uint, entryChannel <-chan []byte, exitDataChannel *webrtc.DataChannel) error {
	return serveThroughClient(TCP, port, entryChannel, exitDataChannel)
}
