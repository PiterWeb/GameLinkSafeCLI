package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/pion/webrtc/v3"
)

func serveThroughClient(protocol, port uint, proxyChan <-chan []byte, endConnChan <-chan uint8, exitDataChannel *webrtc.DataChannel) error {

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

	var pipeCountMutex sync.Mutex
	pipeArr := make([]dataPipe, 255)
	log.Println("Initialized pipe array with size:", len(pipeArr))
	pipeCount := uint8(0)

	go func() {
		for id := range endConnChan {
			log.Printf("Received end connection signal for ID(%d)\n", id)
			pipeArr[id].writer.Close() // Close the writer to signal end of connection
		}
	}()

	go func () {
		for data := range proxyChan {
						
			id := data[0]
			log.Printf("Received data for ID(%d) with len: %d \n", id, len(data)-1)

			if pipeArr[id].reader == nil || pipeArr[id].writer == nil {
				log.Printf("No pipe found for ID(%d), skipping data write\n", id)
				continue
			}

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
			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				n, err := conn.Read(buf)

				if err != nil {
					log.Println("Error reading from connection:", err)
					log.Printf("Closing pipe for ID(%d)\n", id)
					pipeArr[id].writer.Close()
					return
				}

				log.Printf("Read %d bytes from tcp connection for ID(%d)\n", n, id)

				data := make([]byte, n)
				copy(data, buf[:n])

				data = append([]byte{byte(id)}, data...) // Prepend the ID to the data

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
