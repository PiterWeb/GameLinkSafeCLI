package proxy

import (
	"io"

	"github.com/pion/webrtc/v3"
)

// SendThroughUDP sends data through a UDP connection (used for host logic)
func SendThroughUDP(port uint, proxyPipeReader *io.PipeReader, exitChannel *webrtc.DataChannel) error {
	return sendThroughHost(UDP, port, proxyPipeReader, exitChannel)
}

// ServeThroughUDP serves as a UDP server (used for client logic)
func ServeThroughUDP(port uint, proxyPipeReader *io.PipeReader, exitChannel *webrtc.DataChannel) error {
	return serveThroughClient(UDP, port, proxyPipeReader, exitChannel)
}
