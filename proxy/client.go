package proxy

import (
	"context"
	"io"
	"log"
	"net"

	"github.com/pion/datachannel"
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

func serveThroughClientTCP(port uint, dataChannel datachannel.ReadWriteCloser) error {

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

	// ctx, cancelCtx := context.WithCancel(context.Background()) 
	
	for {
		conn, err := listener.Accept()
		
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// cancelCtx()

		ctx, cancelCtx := context.WithCancel(context.Background()) 
		
		dataChannelR := NewReader(ctx, dataChannel)
		dataChannelW := NewWriter(ctx, dataChannel)

		go func() {
			if _, err = io.Copy(dataChannelW, conn); err != nil {
				cancelCtx()
			}
		}()
		
		go func() {
			if _, err = io.Copy(conn, dataChannelR); err != nil {
				cancelCtx()
			}
		}()

	}

}
