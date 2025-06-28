package webrtc

import "github.com/pion/webrtc/v3"

var defaultIceServers = []webrtc.ICEServer{
	{
		URLs: []string{"stun:stun.l.google.com:19305", "stun:stun.l.google.com:19302", "stun:stun.ipfire.org:3478"},
	},
}
