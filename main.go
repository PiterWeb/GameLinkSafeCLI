package main

import (
	"flag"
	"fmt"
	"gamelinksafecli/config"
	"gamelinksafecli/proxy"
	"gamelinksafecli/webrtc"
	"io"
	"io/fs"
	"log"
	"os"
)

const (
	defaultPort     uint   = 8080
	defaultProtocol string = "tcp"
)

func main() {

	log.SetOutput(io.Discard)
	
	rolPtr := flag.String("role", "host", "Role of the application (host/client)")
	portPtr := flag.Uint("port", defaultPort, "Port to listen on")
	protocolPtr := flag.String("protocol", defaultProtocol, "Protocol to use (tcp/udp)")
	configFilePtr := flag.String("config", "servers.yml", "Path to the config file for STUN/TURN urls and credentials")
	flag.BoolFunc("verbose", "Enables logging in stdout", func(string) error {
		log.SetOutput(os.Stdout)
		return nil
	})
	
	flag.Parse()

	rol := *rolPtr
	port := *portPtr
	protocol := *protocolPtr
	configFile := *configFilePtr

	if port < 1 || port > 65535 {
		fmt.Println("Error: Port must be between 1 and 65535")
		return
	}

	if protocol != "tcp" && protocol != "udp" {
		fmt.Println("Error: Protocol must be either 'tcp' or 'udp'")
		return
	}

	if rol != "host" && rol != "client" {
		fmt.Println("Error: Role must be either 'host' or 'client'")
		return
	}
	
	if !fs.ValidPath(configFile) {
		fmt.Println("Error: config file path must be a valid path")
		return
	}
	
	iceServers := config.LoadICEServers(configFile)
	
	fmt.Println("Starting proxy with the following configuration:")
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Protocol: %s\n", protocol)

	var protocolEnum uint

	switch protocol {
	case "tcp":
		protocolEnum = proxy.TCP
	case "udp":
		protocolEnum = proxy.UDP
	}

	if rol == "client" {
		
		err := webrtc.ClientWebrtc(port, protocolEnum, iceServers)

		if err != nil {
			log.Printf("Error starting WebRTC client: %v\n", err)
			return
		}

		return
	}

	if rol == "host" {

		err := webrtc.HostWebrtc(port, protocolEnum, iceServers)

		if err != nil {
			log.Printf("Error starting WebRTC host: %v\n", err)
			return
		}
	}

}
