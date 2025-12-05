package proxy

import (
	"github.com/pion/webrtc/v3"
)

// SendThroughTCP sends data through a TCP connection (used for host logic)
func SendThroughTCP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel) error {
	return sendThroughHostTCP(port, proxyChan, exitDataChannel)
}

// ServeThroughTCP serves as a TCP server (used for client logic)
func ServeThroughTCP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel) error {
	return serveThroughClientTCP(port, proxyChan, exitDataChannel)
}
