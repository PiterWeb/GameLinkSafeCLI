package webrtc

import (
	"fmt"
	"gamelinksafecli/proxy"
	"gamelinksafecli/webrtc/signal"
	"log"
	"strings"

	"github.com/pion/webrtc/v3"
)

func ClientWebrtc(destinationPort uint, finalProtocol uint) error {

	triggerEnd := make(chan error)

	candidates := []webrtc.ICECandidateInit{}

	config := webrtc.Configuration{
		ICEServers: defaultIceServers,
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return err
	}

	defer func() {
		peerConnection.Close()
		close(triggerEnd)
	}()

	endConnChan := make(chan uint8)

	defer close(endConnChan)

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {

		handleDataChannel(destinationPort, finalProtocol, endConnChan,d)
		handleEndConnChannel(d, endConnChan)

	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {

		if c == nil {
			fmt.Println("--- Copy the following line and paste it in the host process to connect ---")
			fmt.Println(signal.SignalEncode(*peerConnection.LocalDescription()) + ";" + signal.SignalEncode(candidates))
			return
		}

		candidates = append(candidates, (*c).ToJSON())

	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		log.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {

			peerConnection.Close()
			triggerEnd <- fmt.Errorf("peer connection failed, closing connection")

		}
	})

	fmt.Println("--- Waiting for the host code ---")

	var offerEncodedWithCandidates string

	_, err = fmt.Scanln(&offerEncodedWithCandidates)

	if err != nil {
		return fmt.Errorf("failed to read offer: %w", err)
	}

	offerEncodedWithCandidatesSplited := strings.Split(offerEncodedWithCandidates, ";")

	offer := webrtc.SessionDescription{}
	_ = signal.SignalDecode(offerEncodedWithCandidatesSplited[0], &offer)

	var receivedCandidates []webrtc.ICECandidateInit

	_ = signal.SignalDecode(offerEncodedWithCandidatesSplited[1], &receivedCandidates)

	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	for _, candidate := range receivedCandidates {
		if err := peerConnection.AddICECandidate(candidate); err != nil {
			log.Println(err)
			continue
		}
	}

	// Create an answer to send to the other process
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return fmt.Errorf("failed to create answer: %w", err)
	}

	// Sets the LocalDescription, and starts our UDP listeners
	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		return fmt.Errorf("failed to set local description: %w", err)
	}

	// Block until error is received or the connection is closed
	<-triggerEnd

	return nil
}

func handleDataChannel(destinationPort uint, finalProtocol uint, endConnChan <-chan uint8, d *webrtc.DataChannel) {

	if d.Label() != "data" {
		return
	}

	proxyChan := make(chan []byte)

	d.OnOpen(func() {

		switch finalProtocol {
		case proxy.UDP:
			_ = proxy.ServeThroughUDP(destinationPort, proxyChan, endConnChan, d)
		case proxy.TCP:
			_ = proxy.ServeThroughTCP(destinationPort, proxyChan, endConnChan, d)
		}

	})

	d.OnMessage(func(msg webrtc.DataChannelMessage) {

		proxyChan <- msg.Data

	})

}

func handleEndConnChannel(d *webrtc.DataChannel, endConnChan chan<- uint8) {

	if d.Label() != "endConn" {
		return
	}

	d.OnMessage(func(msg webrtc.DataChannelMessage) {
		endConnChan <- msg.Data[0]
	})

}