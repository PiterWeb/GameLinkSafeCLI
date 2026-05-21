package proxy

import (
	"github.com/pion/webrtc/v3"
)

// SendThroughTCP sends data through a TCP connection (used for host logic)
func SendThroughTCP(port uint, peerConnection *webrtc.PeerConnection) {
	sendThroughHostTCP(port, peerConnection)
}

// ServeThroughTCP serves as a TCP server (used for client logic)
func ServeThroughTCP(port uint, peerConnection *webrtc.PeerConnection) error {
	return serveThroughClientTCP(port, peerConnection)
}
