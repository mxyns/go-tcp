package filet

import (
	"filet/requests"
	"filet/utils"
	"fmt"
	"net"
	"time"
)

type Client struct {
	*utils.Address
	Socket *net.Conn
}

func (c *Client) Start(timeout string) *Client {

	fmt.Printf("Trying to connect to %v://%v... ", c.Proto, c.Address.ToString())
	timeout_duration, err := time.ParseDuration(timeout)
	socket, err := net.DialTimeout(c.Proto, c.Address.ToString(), timeout_duration)
	if err != nil {
		fmt.Printf("Couldn't reach target. Stopping. %v\n", err)
		return nil
	}
	fmt.Println("Connected.")
	c.Socket = &socket

	return c
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

	response, _, _ := utils.SendRequestOn(c.Socket, request)
	return response
}
