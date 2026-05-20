package proxy

import (
	"github.com/pion/datachannel"
)

// SendThroughUDP sends data through a UDP connection (used for host logic)
func SendThroughUDP(port uint, dataChannel datachannel.ReadWriteCloser) error {
	return sendThroughHostUDP(port, dataChannel)
}

// ServeThroughUDP serves as a UDP server (used for client logic)
func ServeThroughUDP(port uint, dataChannel datachannel.ReadWriteCloser) error {
	return serveThroughClientUDP(port, dataChannel)
}
