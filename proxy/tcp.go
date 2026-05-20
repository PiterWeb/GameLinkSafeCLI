package proxy

import (
	"github.com/pion/datachannel"
)

// SendThroughTCP sends data through a TCP connection (used for host logic)
func SendThroughTCP(port uint, dataChannel datachannel.ReadWriteCloser) error {
	return sendThroughHostTCP(port, dataChannel)
}

// ServeThroughTCP serves as a TCP server (used for client logic)
func ServeThroughTCP(port uint, dataChannel datachannel.ReadWriteCloser) error {
	return serveThroughClientTCP(port, dataChannel)
}
