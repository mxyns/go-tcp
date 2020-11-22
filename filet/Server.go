package filet

// TODO make all import path relative so that it can be imported from github using go -get without having to move the library to a different folder
import (
	"github.com/mxyns/go-tcp/filet/requests"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

const MAX_REQUEST_ID = 255

type Server struct {
	*Address
	Clients          []*net.Conn
	ConnectionWaiter *sync.WaitGroup
	RequestHandler   func(client *net.Conn, request *requests.Request)
}

func (s *Server) Start() {

	listener, err := net.Listen(s.Proto, s.Address.ToString())
	if err != nil {
		log.WithFields(log.Fields{
			"protocol": s.Proto,
			"address":  s.Address.ToString(),
			"error":    err,
		}).Error("Error while trying to open server")
		panic(err)
	}
	defer listener.Close()
	defer s.ConnectionWaiter.Done()

	log.WithFields(log.Fields{
		"protocol": s.Proto,
		"address":  s.Address.ToString(),
	}).Debug("Server running")

	for {
		log.Debug("Waiting for connections...")
		socket, err := listener.Accept()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't accept incoming connection")
		}

		s.ConnectionWaiter.Add(1)

		s.Clients = append(s.Clients, &socket)
		go func() {
			log.WithFields(log.Fields{
				"address": socket.RemoteAddr(),
			}).Debug("Client connected")
			defer socket.Close()
			s.handleClient(&socket)
			s.ConnectionWaiter.Done()
		}()
	}
}
func (s *Server) Close() {

	log.WithFields(log.Fields{"clients": s.Clients}).Debug("Closing server. Closing connections")
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
			log.WithFields(log.Fields{
				"name": (*received).Name(),
			}).Info("Deserialized a request")
			if s.RequestHandler != nil {
				s.RequestHandler(socket, received)
			} else {
				s.Close()
				panic("server doesn't have RequestHandler. stopping server")
			}
		} else if err_id <= MAX_REQUEST_ID {
			log.WithFields(log.Fields{
				"id": err_id,
			}).Error("Error : unknown id")
		} else {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Couldn't use socket. Connection must be closed.")
			break
		}
	}
}
