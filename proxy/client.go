package proxy

import (
	"fmt"
	"log"
	"net"

	"github.com/pion/webrtc/v3"
)

func serveThroughClient(protocol, port uint, entryChannel <-chan []byte, exitDataChannel *webrtc.DataChannel) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	var listener net.Listener
	var err error

	switch protocol {
	case UDP:
		listener, err = net.Listen("udp", addr)
	case TCP:
		listener, err = net.Listen("tcp", addr)
	default:
		listener, err = nil, fmt.Errorf("invalid protocol")
	}

	if err != nil {
		log.Println("Error starting listener:", err)
		return err
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		go func() {

			for buf := range entryChannel {

				_, err = conn.Write(buf)
				if err != nil {
					log.Println("Error writing to connection:", err)
					conn.Close()
					break
				}
			}
		}()

		go func() {
			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				_, err := conn.Read(buf)

				if err != nil {
					log.Println("Error reading from connection:", err)
					conn.Close()
					break
				}

				exitDataChannel.Send(buf)
			}
		}()

	}

	return nil

}
