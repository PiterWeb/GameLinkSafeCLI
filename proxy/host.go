package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"
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

		var wgClose sync.WaitGroup

		wgClose.Add(2)

		go func() {

			defer wgClose.Done()

			for {

				_, err := io.Copy(conn, proxyPipeReader) // Copy data from the pipe to the connection
				
				if err != nil {
					log.Println("Error writing data to connection:", err)
					conn.Close()
					return
				}

			}

		}()

		go func() {

			defer wgClose.Done()

			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				n, err := conn.Read(buf)
				
				if err != nil {
					log.Println("Error reading from connection:", err)
					// conn.Close()
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

		wgClose.Wait() // Wait for both goroutines to finish

	}
	
}
