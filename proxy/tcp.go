package proxy

import (
	"github.com/pion/webrtc/v3"
)

// SendThroughTCP sends data through a TCP connection (used for host logic)
func SendThroughTCP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel, endConnChannel *webrtc.DataChannel) error {
	return sendThroughHost(TCP, port, proxyChan, exitDataChannel, endConnChannel)
}

// ServeThroughTCP serves as a TCP server (used for client logic)
func ServeThroughTCP(port uint, proxyChan <-chan []byte, endConnChan <-chan uint8, exitDataChannel *webrtc.DataChannel) error {
	return serveThroughClientTCP(port, proxyChan, endConnChan, exitDataChannel)
}
