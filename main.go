package main

import (
	"fmt"
	"log"
	"net"
)

type Message struct {
	from    string
	payload []byte
}
type Server struct {
	listenAddr string
	ln         net.Listener
	quitch     chan struct{}
	msgch      chan Message
	clients    map[net.Addr]net.Conn
}

func NewServer(listerAddr string) *Server {
	return &Server{
		listenAddr: listerAddr,
		quitch:     make(chan struct{}),
		msgch:      make(chan Message, 10),
		clients:    make(map[net.Addr]net.Conn),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return err
	}
	defer ln.Close()
	s.ln = ln

	fmt.Println("tcp server started...")
	go s.acceptLoop()

	<-s.quitch
	close(s.msgch)
	return nil
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.ln.Accept()
		s.clients[conn.RemoteAddr()] = conn
		if err != nil {
			fmt.Println("accept error:", err.Error())
			continue
		}
		fmt.Println("new connection to the server: ", conn.RemoteAddr().String())
		go s.readLoop(conn)
	}
}

func (s *Server) readLoop(conn net.Conn) {
	defer conn.Close()
	buff := make([]byte, 2048)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			fmt.Println("read error:", err.Error())
			continue
		}

		s.msgch <- Message{from: conn.RemoteAddr().String(), payload: buff[:n]}

		for _, v := range s.clients {
			v.Write([]byte(fmt.Sprintf("new message received from: %s\n", conn.RemoteAddr().String())))
		}
	}
}

func main() {
	server := NewServer(":3001")

	go func() {
		var msg Message
		for {
			msg = <-server.msgch
			fmt.Printf("received message from:%s: %s \n", msg.from, string(msg.payload))
		}
	}()
	log.Fatal(server.Start())
}
