package fil

import (
	"fil/endpoints"
	"fil/requests"
	"fmt"
	"net"
)

type Client struct {
	*endpoints.Address
	Socket *net.Conn
}

func (c *Client) Start() *Client {

	fmt.Printf("Trying to connect to %v://%v", c.Proto, c.Address.ToString())
	socket, _ := net.Dial(c.Proto, c.Address.ToString())
	fmt.Println("Connected.")
	// GET filters list
	c.Socket = &socket

	return c
}
func (c *Client) Close() {

	fmt.Printf("Closing Client\n")
	err := (*c.Socket).Close()
	if err != nil {
		fmt.Printf("Closed Client socket, caused : %v", err)
	}
}
func (c *Client) Send(request requests.Request) *requests.Request {

	return endpoints.SendRequestOn(c.Socket, request)
}