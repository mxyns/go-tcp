package filet

// TODO make all import path relative so that it can be imported from github using go -get without having to move the library to a different folder
import (
	"github.com/mxyns/go-tcp/filet/requests"
	log "github.com/sirupsen/logrus"
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
		log.Error("Error while trying to open %v server on %v : %v", s.Proto, s.Address.ToString(), err)
		panic(err)
	}
	defer listener.Close()
	defer s.ConnectionWaiter.Done()

	log.Debug("Server running on %v://%v.\n", s.Proto, s.Address.ToString())

	for {
		log.Debug("Waiting for connections...")
		socket, err := listener.Accept()
		if err != nil {
			log.Error("Couldn't accept incoming connection : %v", err)
		}

		s.ConnectionWaiter.Add(1)

		s.Clients = append(s.Clients, &socket)
		go func() {
			log.Debug("Client connected : %v\n", socket.RemoteAddr())
			defer socket.Close()
			s.handleClient(&socket)
			s.ConnectionWaiter.Done()
		}()
	}
}
func (s *Server) Close() {

	log.Debug("Closing server. Closing connections : %v\n", s.Clients)
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
			log.Info("Deserialized a [%v] Request\n", (*received).Name())
			if s.RequestHandler != nil {
				s.RequestHandler(socket, received)
			} else {
				s.Close()
				panic("server doesn't have RequestHandler. stopping server")
			}
		} else if err_id <= 255 {
			log.Error("Error : unknown id %v\n", err_id)
		} else {
			log.Error("Couldn't use socket. Connection must be closed. %v\n", err)
			break
		}
	}
}
