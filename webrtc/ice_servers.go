package webrtc

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pion/webrtc/v3"
)

const StunServersURL string = "https://raw.githubusercontent.com/pradt2/always-online-stun/master/valid_hosts.txt"

func getURLsDefaultIceServers() []string {
	
	resp, err := http.Get(StunServersURL)
	
	if err != nil {
		return []string{}
	}
	
	body, err := io.ReadAll(resp.Body)
	
	stunServers := strings.Split(strings.TrimSpace(string(body)), "\n")
		
	for i := range len(stunServers) {
		stunServers[i] = fmt.Sprintf("stun:%s", stunServers[i])
	}
	
	if len(stunServers) < 4 {
		return stunServers
	}
		
	return stunServers[0:3]
}

var defaultIceServers = []webrtc.ICEServer{
	{
		URLs: getURLsDefaultIceServers(),
	},
}
