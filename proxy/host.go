package proxy

import (
	"fmt"
	"log"
	"net"
	"sync"

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

	var netConnMapMutex sync.Mutex // Mutex to protect access to netConnMap
	netConnMap := make(map[uint8]net.Conn, 255)

	for data := range proxyChan {

		id := data[0] // Assuming the first byte is the ID

		log.Printf("Received data from WebRTC for ID(%d) with len: %d \n", id, len(data)-1)

		netConnMapMutex.Lock() // Lock the map for safe concurrent access
		conn, exists := netConnMap[id]
		netConnMapMutex.Unlock() // Unlock the map after checking

		if !exists {
			
			var err error

			conn, err = net.Dial(network, addr)
			if err != nil {
				log.Println("Error connecting to host:", err)
				log.Printf("Closing connection for ID(%d)\n", id)
				netConnMapMutex.Lock() // Lock the map for safe concurrent access
				delete(netConnMap, id) // Remove the connection from the map
				netConnMapMutex.Unlock() // Unlock the map after deleting
				endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
				continue
			}

			netConnMapMutex.Lock() // Lock the map for safe concurrent access
			netConnMap[id] = conn // Store the connection in the map
			netConnMapMutex.Unlock() // Unlock the map after storing

			log.Printf("Established new connection for ID(%d)\n", id)

		}
	
		// Only start reading from the connection if it doesn't already exist
		// This prevents multiple goroutines from reading from the same connection
		// and ensures that we only read once per connection
		if !exists {
			go func() {
				buf := make([]byte, 65507) // Maximum UDP packet size
				for {
					n, err := conn.Read(buf)
					
					if err != nil {
						log.Println("Error reading from connection:", err)
						log.Printf("Closing connection for ID(%d)", id)
						netConnMapMutex.Lock() // Lock the map for safe concurrent access
						delete(netConnMap, id) // Remove the connection from the map
						netConnMapMutex.Unlock() // Unlock the map after deleting
						endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
						return
					}
					
					log.Printf("Read %d bytes from connection for ID(%d)\n", n, id)
					
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

		n, err := conn.Write(data[1:]) // Write data excluding the ID byte
		if err != nil {
			log.Println("Error writing data to connection:", err)
			log.Printf("Closing connection for ID(%d)\n", id)
			netConnMapMutex.Lock() // Lock the map for safe concurrent access
			delete(netConnMap, id) // Remove the connection from the map
			netConnMapMutex.Unlock() // Unlock the map after deleting
			endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
			continue
		}
		
		log.Printf("Wrote %d bytes to connection for ID(%d)\n", n, id)
		
	}
	
	return nil
	
}
