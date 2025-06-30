package config

import (
	"fmt"
	"os"

	"github.com/pion/webrtc/v3"
	"gopkg.in/yaml.v3"
)

type webRTCConfig = webrtc.Configuration

func LoadICEServers(config_path string) []webrtc.ICEServer {
	
	_, err := os.Stat(config_path)
	
	if err != nil {
		return []webrtc.ICEServer{}
	}
	
	fileBytes, err := os.ReadFile(config_path)
	
	if err != nil {
		return []webrtc.ICEServer{}
	}
	
	iceServers := new(webRTCConfig)
	
	err = yaml.Unmarshal(fileBytes, iceServers)
	
	if err != nil {
		fmt.Println("Error: config file cannot be parsed as yaml")
	}
	
	fmt.Println("Loaded config file successfully")
	
	return iceServers.ICEServers
	
}