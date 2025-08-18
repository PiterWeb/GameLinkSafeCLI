package proxy

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v3"
)

func handleIncomingUDPPackets(conn net.Conn, connExists *atomic.Bool, exitDataChannel *webrtc.DataChannel) {
	buf := make([]byte, 65507) // Maximum UDP packet size
	for {
		n, err := conn.Read(buf)
		
		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

		if err != nil {
			log.Println("Error reading from connection:", err)
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("Closing connection")
				conn.Close()
			}
			connExists.Store(false)
			return
		}

		log.Printf("Read %d bytes from connection\n", n)
		
		log.Println("Sending data through webrtc:")
		err = exitDataChannel.Send(buf[:n])
		
		if err != nil {
			log.Println("Error sending data through webrtc:", err)
		}

		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	
	}
}

func sendThroughHostUDP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	
	dialer := net.Dialer{
		LocalAddr: &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 0, // Setting port to 0 will make the OS choose a random available port
		},
	}
	
	network := "udp"

	var conn net.Conn
	connExists := new(atomic.Bool)
	connExists.Store(false) 

	for data := range proxyChan {

		log.Printf("Received data from WebRTC with len: %d \n", len(data)-1)

		if !connExists.Load() {
			
			var err error

			conn, err = dialer.Dial(network, addr)

			if err != nil {
				log.Println("Error connecting to host:", err)
				continue
			}

			log.Printf("Established new connection\n")

			// Only start reading from the connection if it doesn't already exist
			// This prevents multiple goroutines from reading from the same connection
			// and ensures that we only read once per connection
			go handleIncomingUDPPackets(conn, connExists, exitDataChannel)

		}

		connExists.Store(true)

		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))

		n, err := conn.Write(data)
		if err != nil {
			log.Println("Error writing data to connection:", err)
			if !errors.Is(err, net.ErrClosed) {
				log.Printf("Closing connection")
				conn.Close()
			}
			continue
		}
		
		log.Printf("Wrote %d bytes to connection\n", n)

		_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
		
	}
	
	return nil
	
}

func sendThroughHostTCP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel, endConnChannel *webrtc.DataChannel) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	localAddr := &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"), // Use your desired local IP address
			Port: 0, // Setting port to 0 will make the OS choose a random available port
	}
	
	dialer := net.Dialer{
		LocalAddr: localAddr,
	}
	
	network := "tcp"

	var netConnMapMutex sync.Mutex // Mutex to protect access to netConnMap
	netConnMap := make(map[uint8]net.Conn, math.MaxUint8)

	for data := range proxyChan {

		id := data[0] // Assuming the first byte is the ID

		log.Printf("Received data from WebRTC for ID(%d) with len: %d \n", id, len(data)-1)

		netConnMapMutex.Lock() // Lock the map for safe concurrent access
		conn, exists := netConnMap[id]
		netConnMapMutex.Unlock() // Unlock the map after checking

		if !exists {
			
			var err error

			conn, err = dialer.Dial(network, addr)
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
