package proxy

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/pion/webrtc/v3"
)

func handleIncomingPackets(conn net.Conn, connExists *atomic.Bool, exitDataChannel *webrtc.DataChannel) {
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
			go handleIncomingPackets(conn, connExists, exitDataChannel)

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

func sendThroughHostTCP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	localAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"), // Use your desired local IP address
		Port: 0,                        // Setting port to 0 will make the OS choose a random available port
	}

	dialer := net.Dialer{
		LocalAddr: localAddr,
	}

	network := "tcp"

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
			go handleIncomingPackets(conn, connExists, exitDataChannel)

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
