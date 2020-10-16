package fil

import (
	"fil/endpoints"
	"fmt"
	"net"
	"sync"
)

type Server struct {

	*endpoints.Address
	Clients []*net.Conn
	ConnectionWaiter *sync.WaitGroup
}

func (s *Server) Start() {

	listener, err := net.Listen(s.Proto, s.Address.ToString())
	if err != nil {
		fmt.Printf("Error while trying to open %v server on %v : %v", s.Proto, s.Address.ToString(), err)
		panic(err)
	}
	defer listener.Close()
	defer s.ConnectionWaiter.Done()

	fmt.Printf("Server running on %v://%v.\n", s.Proto, s.Address.ToString())

	for {
		fmt.Println("Waiting for connections...")
		socket, err := listener.Accept()
		if err != nil {
			fmt.Printf("Couldn't accept incoming connection : %v", err)
		}

		s.ConnectionWaiter.Add(1)

		s.Clients = append(s.Clients, &socket)
		go func() {
			fmt.Printf("Client connected : %v\n", socket.RemoteAddr())
			defer socket.Close()
			s.handleClient(&socket)
			s.ConnectionWaiter.Done()
		}()
	}
}
func (s *Server) Close() {

	fmt.Printf("Closing server. Closing connections : %v\n", s.Clients)
	for i := range s.Clients {
		if s.Clients[i] != nil {
			(*s.Clients[i]).Close()
		}
	}
}

func (s *Server) handleClient(socket *net.Conn) {

	// SEND filters list
	for {

		received := endpoints.Await(socket)
		id := (*received).Info().Id

		if received != nil {
			fmt.Printf("Deserialized a [%v] Request\n", (*received).Name())
			if (*received).NeedsResponse() {
				endpoints.SendRequestOn(socket, (*received).GetResult())
			}
		} else {
			fmt.Printf("Error : unknown id %v\n", id)
		}
	}
}
