package filet

import (
	"github.com/mxyns/go-tcp/filet/requests"
	log "github.com/sirupsen/logrus"
	"net"
	"time"
)

type Client struct {
	*Address
	Socket *net.Conn
}

func (c *Client) Start(timeout string) (*Client, error) {

	log.WithFields(log.Fields{
		"protocol": c.Proto,
		"address":  c.Address.ToString(),
	}).Debug("Trying to connect to server.")
	timeout_duration, err := time.ParseDuration(timeout)
	socket, err := net.DialTimeout(c.Proto, c.Address.ToString(), timeout_duration)
	if err != nil {
		return nil, err
	}
	log.Debug("Connected.")
	c.Socket = &socket

	return c, nil
}
func (c *Client) Close() {

	log.Debug("Closing Client")
	if c.Socket == nil {
		log.Error("Client already closed.")
		return
	}

	err := (*c.Socket).Close()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Client socket closed")
	}
}
func (c *Client) Send(request requests.Request) *requests.Request {

	response, _, _ := requests.SendRequestOn(c.Socket, &request)
	return response
}
