package proxy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
	// "time"

	"github.com/pion/datachannel"
	"github.com/pion/webrtc/v4"
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

		// conn.SetDeadline(time.Now().Add(time.Minute))
		
		d, err := peerConnection.CreateDataChannel(fmt.Sprintf("tcp%d", connCounter.Add(1)), &webrtc.DataChannelInit{
			Ordered: &ordered,
		})

		if err != nil {
			log.Println("Error creating datachannel: ", err)
			conn.Close()
			return err
		}

		d.OnOpen(func() {
			defer conn.Close()
			defer d.Close()
			
			dataChannel, err := d.Detach()

			if err != nil {
				log.Println("Error detach datachannel: ", err)
				return
			}

			defer dataChannel.Close()
			
			log.Println("Established new connection")
			
			var wg sync.WaitGroup
    		wg.Add(2)

		    go func() {
		        defer wg.Done()
		        io.Copy(dataChannel, bufio.NewReader(conn))
		        d.Close()
				dataChannel.Close()
		    }()
		    go func() {
		        defer wg.Done()
		        io.Copy(conn, bufio.NewReader(dataChannel))
		        conn.Close()
		    }()

			wg.Wait()	

		})
	}

}