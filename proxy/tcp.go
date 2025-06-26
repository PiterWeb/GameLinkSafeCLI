package proxy

import (
	"io"

	"github.com/pion/webrtc/v3"
)

// SendThroughTCP sends data through a TCP connection (used for host logic)
func SendThroughTCP(port uint, proxyPipeReader *io.PipeReader, exitDataChannel *webrtc.DataChannel) error {
	return sendThroughHost(TCP, port, proxyPipeReader, exitDataChannel)
}

// ServeThroughTCP serves as a TCP server (used for client logic)
func ServeThroughTCP(port uint,proxyPipeReader *io.PipeReader, exitDataChannel *webrtc.DataChannel) error {
	return serveThroughClient(TCP, port, proxyPipeReader, exitDataChannel)
}
