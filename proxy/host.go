package proxy

import (
	"fmt"
	"log"
	"net"

	"github.com/pion/webrtc/v3"
)

func sendThroughHost(protocol, port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel, endConnChannel *webrtc.DataChannel) error {

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

	netConnMap := make(map[uint8]net.Conn, 255)

	for data := range proxyChan {

		id := data[0] // Assuming the first byte is the ID

		log.Printf("Received data for ID(%d) with len: %d \n", id, len(data)-1)

		conn, exists := netConnMap[id]

		if exists {
			
			log.Printf("Using existing connection for ID(%d)\n", id)
			n, err := conn.Write(data[1:]) // Write data excluding the ID byte
			if err != nil {
				log.Println("Error writing data to connection:", err)
			}

			log.Printf("Wrote %d bytes to connection for ID(%d)\n", n, id)

			delete(netConnMap, id) // Remove the connection from the map
			endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
			continue

		}

		var err error
		conn, err = net.Dial(network, addr)
		if err != nil {
			log.Println("Error connecting to host:", err)
			log.Println("Closing connection for ID:", id)
			endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
			continue
		}

		log.Println("Established new connection for ID:", id)

		netConnMap[id] = conn // Store the connection in the map
	
		go func() {
			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				n, err := conn.Read(buf)
				
				if err != nil {
					log.Println("Error reading from connection:", err)
					log.Println("Closing connection for ID:", id)
					delete(netConnMap, id) // Remove the connection from the map
					endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
					return
				}
				
				// Save buffered data to a new slice
				data := make([]byte, n)
				copy(data, buf[:n])

				data = append([]byte{byte(id)}, data...) // Prepend the ID to the data

				log.Println("Sending data through webrtc for ID:", id)
				err = exitDataChannel.Send(data)
				
				if err != nil {
					log.Println("Error sending data through webrtc:", err)
				}
			
			}
		}()

	}

	return nil

}
