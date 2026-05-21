package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v3"
)

func serveThroughClientUDP(port uint, dataChannel datachannel.ReadWriteCloser) error {

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
	
	go io.Copy(dataChannel, listener)
	_, err = io.Copy(listener, dataChannel)

	return err
}

func serveThroughClientTCP(port uint, peerConnection *webrtc.PeerConnection) error {

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

	var connCounter atomic.Uint64
	
	ordered := true
	
	for {
		conn, err := listener.Accept()
		
		if err != nil {
			log.Println("Error accepting connection: ", err)
			continue
		}

		conn.SetDeadline(time.Now().Add(time.Minute))
		
		datachannel, err := peerConnection.CreateDataChannel(fmt.Sprintf("tcp%d", connCounter.Add(1)), &webrtc.DataChannelInit{
			Ordered: &ordered,
		})

		if err != nil {
			log.Println("Error creating datachannel: ", err)
			conn.Close()
			return err
		}

		datachannel.OnOpen(func() {
			defer conn.Close()
			defer datachannel.Close()
			
			d, err := datachannel.Detach()

			if err != nil {
				log.Println("Error detach datachannel: ", err)
				return
			}
			
			go io.Copy(d, bufio.NewReader(conn))
			io.Copy(conn, bufio.NewReader(d))

		})
	}

}