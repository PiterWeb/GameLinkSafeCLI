package proxy

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	"github.com/pion/datachannel"
)

func sendThroughHostUDP(port uint, dataChannel datachannel.ReadWriteCloser) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	dialer := net.Dialer{
		LocalAddr: &net.UDPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 0, // Setting port to 0 will make the OS choose a random available port
		},
	}

	const network = "udp"

	for {
		
		conn, err := dialer.Dial(network, addr)

		if err != nil {
			log.Println("Error connecting to host:", err)
			continue
		}

		log.Printf("Established new connection\n")
		
		go io.Copy(dataChannel, conn)
		io.Copy(conn, dataChannel)
	}

}

func sendThroughHostTCP(port uint, dataChannel datachannel.ReadWriteCloser) error {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	localAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,                        // Setting port to 0 will make the OS choose a random available port
	}

	dialer := net.Dialer{
		LocalAddr: localAddr,
	}

	const network = "tcp"
	
	for {
		
		conn, err := dialer.Dial(network, addr)

		if err != nil {
			log.Println("Error connecting to host:", err)
			time.Sleep(time.Second)
			continue
		}

		log.Printf("Established new connection\n")

		go io.Copy(dataChannel, conn)
		
		io.Copy(conn, dataChannel)
	}

}
