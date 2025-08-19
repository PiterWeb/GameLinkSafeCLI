package proxy

import (
	"io"
	"log"
	"math"
	"net"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

func serveThroughClientUDP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel) error {

	addr := net.UDPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: int(port),
	}

	listener, err := net.ListenUDP("udp", &addr)

	if err != nil {
		log.Println("Error starting listener:", err)
		return err
	}

	defer listener.Close()
	var remoteAddr net.Addr

	go func() {
		for data := range proxyChan {

			n, err := listener.WriteTo(data, remoteAddr)
			
			if err != nil {
				log.Println("Error writing data to connection:", err)
			}
			
			log.Printf("Finished writing %d bytes to connection\n", n)

		}
	}()

	buf := make([]byte, 0, 65507) // Maximum UDP packet size
	for {

		var n int
		n, remoteAddr, err = listener.ReadFrom(buf[:cap(buf)])

		if err != nil {
			log.Println("Error reading from connection:", err)
			continue
		}

		log.Printf("Read %d bytes from udp connection\n", n)
		exitDataChannel.Send(buf[:n])

	}
}

func serveThroughClientTCP(port uint, proxyChan <-chan []byte, endConnChan <-chan uint8, exitDataChannel *webrtc.DataChannel) error {

	addr := net.TCPAddr{
		IP: net.ParseIP("127.0.0.1"),
		Port: int(port),
	}

	listener, err := net.ListenTCP("tcp", &addr)

	if err != nil {
		log.Println("Error starting listener:", err)
		return err
	}

	defer listener.Close()

	var pipeCountMutex sync.Mutex
	pipeArr := make([]dataPipe, math.MaxUint8)
	log.Println("Initialized pipe array with size:", len(pipeArr))
	pipeCount := uint8(0)

	go func() {
		for id := range endConnChan {
			go func(id uint8) {
				log.Printf("Received end connection signal for ID(%d)\n", id)
				time.Sleep(1 * time.Second) // Wait for 1 second before closing the pipe
				pipeArr[id].writer.Close()  // Close the writer to signal end of connection
			}(id)
		}
	}()

	go func() {
		for data := range proxyChan {

			id := data[0]
			log.Printf("Received data from WebRTC for ID(%d) with len: %d \n", id, len(data)-1)

			n, err := pipeArr[id].writer.Write(data[1:])

			if err != nil {
				log.Printf("Error writing to pipe with ID(%d): %v\n", id, err)
				continue
			}

			log.Printf("Wrote %d bytes to pipe for ID(%d)\n", n, id)

		}
	}()

	for {
		conn, err := listener.Accept()

		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		pipeCountMutex.Lock()
		id := pipeCount
		pipeCountMutex.Unlock()

		pipeArr[id].reader, pipeArr[id].writer = io.Pipe()
		log.Printf("Created new pipe for ID(%d)\n", id)

		pipeCountMutex.Lock()
		pipeCount++
		log.Println("Incremented pipe count to:", pipeCount)
		if pipeCount >= uint8(len(pipeArr)) {
			pipeCount = 0 // Reset pipe count if it exceeds the array length
		}
		pipeCountMutex.Unlock()

		go func() {
			for {
				buf := make([]byte, 65507) // Maximum UDP packet size
				n, err := conn.Read(buf)

				if err != nil {
					log.Println("Error reading from connection:", err)
					log.Printf("Closing pipe for ID(%d)\n", id)
					pipeArr[id].writer.Close()
					return
				}

				log.Printf("Read %d bytes from tcp connection for ID(%d)\n", n, id)

				data := append([]byte{id}, buf[:n]...) // Prepend the ID to the data

				log.Printf("Sending data through WebRTC for ID(%d)\n", id)
				exitDataChannel.Send(data)
			}
		}()

		go func() {
			n, err := io.Copy(conn, pipeArr[id].reader) // Copy data from the pipe to the connection
			if err != nil {
				log.Println("Error writing data to connection:", err)
			}

			log.Printf("Finished writing %d bytes to connection for ID(%d)\n", n, id)
		}()

	}

}
