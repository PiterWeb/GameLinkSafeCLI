package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"time"

	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
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

		log.Println("Established new connection")
		
		go io.Copy(dataChannel, conn)
		io.Copy(conn, dataChannel)
	}

}

func sendThroughHostTCP(port uint, peerConnection *webrtc.PeerConnection) {

	addr := fmt.Sprintf("127.0.0.1:%d", port)

	localAddr := &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,                        // Setting port to 0 will make the OS choose a random available port
	}

	dialer := net.Dialer{
		LocalAddr: localAddr,
	}

	const network = "tcp"

	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {

		if !strings.Contains(d.Label(), "tcp") {
			return
		} 
		
		d.OnOpen(func() {

			dataChannel, err := d.Detach()

			if err != nil {
				log.Println("Error detach datachannel: ", err)
				return
			}

			defer dataChannel.Close()
			
			conn, err := dialer.Dial(network, addr)

			if err != nil {
				log.Println("Error connecting to host, closing datachannel:", err)
				return
			}

			conn.SetDeadline(time.Now().Add(time.Minute))
			
			defer conn.Close()
			
			log.Println("Established new connection")
			
			go io.Copy(dataChannel, bufio.NewReader(conn))
			io.Copy(conn, bufio.NewReader(dataChannel))
		})
	})
}