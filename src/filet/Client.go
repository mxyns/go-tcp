package filet

import (
	"filet/requests"
	"fmt"
	"net"
	"time"
)

type Client struct {
	*Address
	Socket *net.Conn
}

func (c *Client) Start(timeout string) (*Client, error) {

	fmt.Printf("Trying to connect to %v://%v...\n", c.Proto, c.Address.ToString())
	timeout_duration, err := time.ParseDuration(timeout)
	socket, err := net.DialTimeout(c.Proto, c.Address.ToString(), timeout_duration)
	if err != nil {
		return nil, err
	}
	fmt.Println("Connected.")
	c.Socket = &socket

	return c, nil
}
func (c *Client) Close() {

	fmt.Println("Closing Client")
	if c.Socket == nil {
		fmt.Println("Client already closed.")
		return
	}

	err := (*c.Socket).Close()
	if err != nil {
		fmt.Printf("Closed Client socket, caused : %v", err)
	}
}
func (c *Client) Send(request requests.Request) *requests.Request {

	response, _, _ := SendRequestOn(c.Socket, &request)
	return response
}
