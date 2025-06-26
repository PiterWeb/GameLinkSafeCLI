package proxy

import (
	"fmt"
	"log"
	"net"

	"github.com/pion/webrtc/v3"
)

func sendThroughHost(protocol, port uint, entryChannel <-chan []byte, exitDataChannel *webrtc.DataChannel) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	var conn net.Conn
	var err error

	switch protocol {
	case UDP:
		conn, err = net.Dial("udp", addr)
	case TCP:
		conn, err = net.Dial("tcp", addr)
	default:
		conn, err = nil, fmt.Errorf("invalid protocol")
	}

	if err != nil {
		log.Println("Error connecting to host:", err)
		return err
	}

	defer conn.Close()

	go func() {

		for buf := range entryChannel {
			log.Println("Receiving data through webrtc:", len(buf), "bytes")
			_, err := conn.Write(buf)
			if err != nil {
				log.Println("Error writing to connection:", err)
				conn.Close()
				break
			}
		}

	}()

	buf := make([]byte, 65507) // Maximum UDP packet size
	for {
		_, err := conn.Read(buf)

		if err != nil {
			log.Println("Error reading from connection:", err)
			conn.Close()
			break
		}

		log.Println("Sending data through webrtc:", len(buf), "bytes")

		err = exitDataChannel.Send(buf)

		if err != nil {
			log.Println("Error sending data through webrtc:", err)
		}

	}

	return nil

}
