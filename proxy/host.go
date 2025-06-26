package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/pion/webrtc/v3"
)

func sendThroughHost(protocol, port uint, proxyPipeReader *io.PipeReader, exitDataChannel *webrtc.DataChannel) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	var network string

	switch protocol {
	case UDP:	
		network = "udp"
	case TCP:
		network = "tcp"
	default:
		return fmt.Errorf("invalid protocol")
	}

	for {

		conn, err := net.Dial(network, addr)

		if err != nil {
			log.Println("Error connecting to host:", err)
			log.Println("Reconnecting in 20 milliseconds...")
			time.Sleep(20 * time.Millisecond) // Wait before reconnecting
			continue
		}

		go func() {

			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				n, err := conn.Read(buf)
				
				if err != nil {
					log.Println("Error reading from connection:", err)
					return
				}
				
				data := make([]byte, n)
				copy(data, buf[:n])
				err = exitDataChannel.Send(data)
				
				if err != nil {
					log.Println("Error sending data through webrtc:", err)
				}
			
			}
		}()

		_, err = io.Copy(conn, proxyPipeReader) // Copy data from the pipe to the connection
		if err != nil {
			log.Println("Error writing data to connection:", err)
		}

		time.Sleep(20 * time.Millisecond) // Wait before reconnecting

		log.Println("Reconnecting to host...")

	}
	
}
