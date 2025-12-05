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
	buf := make([]byte, 0, 65507) // Maximum UDP packet size
	for {
		n, err := conn.Read(buf[:cap(buf)])
		
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

		log.Printf("Received data from WebRTC with len: %d \n", len(data))

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

			connExists.Store(true)
		}

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

	var netConnMapMutex sync.RWMutex // Mutex to protect access to netConnMap
	netConnMap := make(map[byte]net.Conn, math.MaxUint8)

	for data := range proxyChan {

		id := data[0] // Assuming the first byte is the ID

		log.Printf("Received data from WebRTC for ID(%d) with len: %d \n", id, len(data)-1)

		netConnMapMutex.RLock() // Lock the map for safe concurrent access
		conn, exists := netConnMap[id]
		netConnMapMutex.RUnlock() // Unlock the map after checking

		if !exists {
			
			var err error

			conn, err = dialer.Dial(network, addr)
			if err != nil {
				log.Println("Error connecting to host:", err)
				log.Printf("Closing connection for ID(%d)\n", id)
				endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
				continue
			}

			netConnMapMutex.Lock() // Lock the map for safe concurrent access
			netConnMap[id] = conn // Store the connection in the map
			netConnMapMutex.Unlock() // Unlock the map after storing

			log.Printf("Established new connection for ID(%d)\n", id)

		}

		_ = conn.SetDeadline(time.Now().Add(time.Second * 2))

		// Only start reading from the connection if it doesn't already exist
		// This prevents multiple goroutines from reading from the same connection
		// and ensures that we only read once per connection
		if !exists {
			go func(id byte, conn net.Conn) {
				buf := make([]byte, 65507) // Maximum UDP packet size
				for {
					n, err := conn.Read(buf[:cap(buf)])

					if err != nil {
						log.Println("Error reading from connection:", err)
						log.Printf("Closing connection for ID(%d)", id)
						netConnMapMutex.Lock() // Lock the map for safe concurrent access
						delete(netConnMap, id) // Remove the connection from the map
						netConnMapMutex.Unlock() // Unlock the map after deleting
						endConnChannel.Send([]byte{uint8(id)}) // Notify end of connection
						_ = conn.Close()
						return
					}
					
					log.Printf("Read %d bytes from connection for ID(%d)\n", n, id)
					
					// Save buffered data to a new slice
					data := make([]byte, 1, n+1) // +1 for the ID
					data[0] = id        // Prepend the ID to the data
					data = append(data, buf[:n]...)   // Copy content of buf from position 1 to n+1

					log.Println("Sending data through webrtc for ID:", id)
					err = exitDataChannel.Send(data)
					
					if err != nil {
						log.Println("Error sending data through webrtc:", err)
					}
				
				}
			}(id, conn)
		}

		n, err := conn.Write(data[1:]) // Write data excluding the ID byte
		if err != nil {
			log.Println("Error writing data to connection:", err)
			log.Printf("Closing connection for ID(%d)\n", id)
			netConnMapMutex.Lock() // Lock the map for safe concurrent access
			delete(netConnMap, id) // Remove the connection from the map
			netConnMapMutex.Unlock()        // Unlock the map after deleting
			endConnChannel.Send([]byte{id}) // Notify end of connection
			_ = conn.Close()
			continue
		}
		
		log.Printf("Wrote %d bytes to connection for ID(%d)\n", n, id)
		
	}
	
	return nil
	
}
