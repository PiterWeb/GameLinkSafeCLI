package webrtc

import (
	"fmt"
	"gamelinksafecli/proxy"
	"gamelinksafecli/webrtc/signal"
	"log"
	"strings"

	"github.com/pion/webrtc/v3"
)

func HostWebrtc(port uint, protocol uint) error {

	triggerEnd := make(chan error)
	defer close(triggerEnd)

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

	endConnChannel, err := peerConnection.CreateDataChannel("endConn", &webrtc.DataChannelInit{})

	if err != nil {
		return err
	}

	ordered := true

	// If the protocol is UDP, we need to set the data channel to unordered
	if protocol == proxy.UDP {
		ordered = false
	}

	dataChannel, err := peerConnection.CreateDataChannel("data", &webrtc.DataChannelInit{
		Ordered: &ordered,
	})

	if err != nil {
		return err
	}

	proxyChan := make(chan []byte)
	defer close(proxyChan)

	// Open the data channel and select the protocol to send data
	dataChannel.OnOpen(func() {

		switch protocol {
		case proxy.UDP:
			_ = proxy.SendThroughUDP(port, proxyChan, dataChannel, endConnChannel)
		case proxy.TCP:
			_ = proxy.SendThroughTCP(port, proxyChan, dataChannel, endConnChannel)
		}

	})

	// Listen for messages on the data channel to send to the proxy channel
	dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) { 
		proxyChan <- msg.Data
	})

	peerConnection.OnICECandidate(func(c *webrtc.ICECandidate) {

		if c == nil {
			fmt.Println("--- Copy the following line and paste it in the client process to connect ---")
			fmt.Println(signal.SignalEncode(*peerConnection.LocalDescription()) + ";" + signal.SignalEncode(candidates))
			fmt.Println("--- Waiting for the client code ---")
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

	offer, err := peerConnection.CreateOffer(nil)

	if err != nil {
		return err
	}

	if err := peerConnection.SetLocalDescription(offer); err != nil {
		return err
	}

	var answerResponse string
	_, err = fmt.Scanln(&answerResponse)

	if err != nil {
		return fmt.Errorf("failed to read answer: %w", err)
	}

	answerEncodedWithCandidatesSplited := strings.Split(answerResponse, ";")

	answer := webrtc.SessionDescription{}

	_ = signal.SignalDecode(answerEncodedWithCandidatesSplited[0], &answer)

	remoteCandidates := []webrtc.ICECandidateInit{}

	_ = signal.SignalDecode(answerEncodedWithCandidatesSplited[1], &remoteCandidates)

	if err := peerConnection.SetRemoteDescription(answer); err != nil {
		return fmt.Errorf("failed to set remote description: %w", err)
	}

	for _, candidate := range remoteCandidates {
		err := peerConnection.AddICECandidate(candidate)

		if err != nil {
			return fmt.Errorf("failed to add ICE candidate: %w", err)
		}
	}

	// Block until cancel by user
	err = <-triggerEnd

	return err

}