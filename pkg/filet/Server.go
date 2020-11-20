package filet

// TODO make all import path relative so that it can be imported from github using go -get without having to move the library to a different folder
import (
	"fmt"
	"github.com/mxyns/go-tcp/filet/requests"
	"net"
	"sync"
)

type Server struct {
	*Address
	Clients          []*net.Conn
	ConnectionWaiter *sync.WaitGroup
	RequestHandler   func(client *net.Conn, request *requests.Request)
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

	for {

		received, err, err_id := requests.Await(socket)

		if err == nil {
			fmt.Printf("Deserialized a [%v] Request\n", (*received).Name())
			if s.RequestHandler != nil {
				s.RequestHandler(socket, received)
			} else {
				s.Close()
				panic("server doesn't have RequestHandler. stopping server")
			}
		} else if err_id <= 255 {
			fmt.Printf("Error : unknown id %v\n", err_id)
		} else {
			fmt.Printf("Couldn't use socket. Connection must be closed. %v\n", err)
			break
		}
	}
}
