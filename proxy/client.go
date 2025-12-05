package proxy

import (
	"log"
	"net"
	"sync/atomic"

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
	remoteAddr := atomic.Pointer[net.Addr]{}

	go func() {
		for data := range proxyChan {

			n, err := listener.WriteTo(data, *remoteAddr.Load())
			
			if err != nil {
				log.Println("Error writing data to connection:", err)
			}
			
			log.Printf("Finished writing %d bytes to connection\n", n)

		}
	}()

	buf := make([]byte, 0, 65507) // Maximum UDP packet size
	for {

		n, tempAddr, err := listener.ReadFrom(buf[:cap(buf)])
		if tempAddr.String() != (*remoteAddr.Load()).String() {
			remoteAddr.Store(&tempAddr)
		}

		if err != nil {
			log.Println("Error reading from connection:", err)
			continue
		}

		log.Printf("Read %d bytes from udp connection\n", n)
		exitDataChannel.Send(buf[:n])

	}
}

func serveThroughClientTCP(port uint, proxyChan <-chan []byte, exitDataChannel *webrtc.DataChannel) error {

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
	conn := atomic.Pointer[net.Conn]{}
	
	go func() {
		for data := range proxyChan {
			
			n, err := (*conn.Load()).Write(data)
			
			if err != nil {
				log.Println("Error writing data to connection:", err)
			}
			
			log.Printf("Finished writing %d bytes to connection\n", n)

		}
	}()
	
	for {
		_conn, err := listener.Accept()
		
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}
		
		oldConn := conn.Swap(&_conn)
		
		if oldConn != nil {
			(*oldConn).Close()
		}
		
		go func() {
			buf := make([]byte, 65507) // Maximum UDP packet size
			for {
				n, err := (*conn.Load()).Read(buf[:cap(buf)])

				if err != nil {
					log.Println("Error reading from connection:", err)
					return
				}

				exitDataChannel.Send(buf[:n])
			}
		}()
		
	}

}
