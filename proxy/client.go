package proxy

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/pion/webrtc/v3"
)

func serveThroughClient(protocol, port uint, proxyPipeReader *io.PipeReader, exitDataChannel *webrtc.DataChannel) error {

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
			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				n, err := conn.Read(buf)

				if err != nil {
					log.Println("Error reading from connection:", err)
					return
				}

				data := make([]byte, n)
				copy(data, buf[:n])

				exitDataChannel.Send(data)
			}
		}()

		_, err = io.Copy(conn, proxyPipeReader) // Copy data from the pipe to the connection
		if err != nil {
			log.Println("Error writing data to connection:", err)
		}

	}

}
