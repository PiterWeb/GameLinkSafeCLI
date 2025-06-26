package main

import (
	"flag"
	"fmt"
	"gamelinksafecli/proxy"
	"gamelinksafecli/webrtc"
)

const (
	defaultPort     uint   = 8080
	defaultProtocol string = "tcp"
)

func main() {

	rolPtr := flag.String("role", "host", "Role of the application (host/client)")
	portPtr := flag.Uint("port", defaultPort, "Port to listen on")
	protocolPtr := flag.String("protocol", defaultProtocol, "Protocol to use (tcp/udp)")

	flag.Parse()

	rol := *rolPtr
	port := *portPtr
	protocol := *protocolPtr

	if port < 1 || port > 65535 {
		fmt.Println("Error: Port must be between 1 and 65535")
		return
	}

	if protocol != "tcp" && protocol != "udp" {
		fmt.Println("Error: Protocol must be either 'tcp' or 'udp'")
		return
	}

	fmt.Println("Starting proxy with the following configuration:")
	fmt.Printf("Port: %d\n", port)
	fmt.Printf("Protocol: %s\n", protocol)

	if rol != "host" && rol != "client" {
		fmt.Println("Error: Role must be either 'host' or 'client'")
		return
	}

	var protocolEnum uint

	switch protocol {
	case "tcp":
		protocolEnum = proxy.TCP
	case "udp":
		protocolEnum = proxy.UDP
	}

	if rol == "client" {
		
		err := webrtc.ClientWebrtc(port, protocolEnum)

		if err != nil {
			fmt.Printf("Error starting WebRTC client: %v\n", err)
			return
		}

		return
	}

	if rol == "host" {

		err := webrtc.HostWebrtc(port, protocolEnum)

		if err != nil {
			fmt.Printf("Error starting WebRTC host: %v\n", err)
			return
		}
	}

}
