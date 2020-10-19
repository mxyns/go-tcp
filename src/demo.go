package main

import (
	"bufio"
	fio "fileio"
	"filet"
	"filet/requests"
	defaultRequests "filet/requests/default"
	"filet/utils"
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
)

func main() {

	runServer := flag.Bool("s", false, "run in server mode")
	address := flag.String("a", "127.0.0.1", "address to host on / connect to")
	proto := flag.String("P", "tcp", "protocol")
	port := uint32(*flag.Uint("p", 8888, "port"))
	timeout := flag.String("t", "10s", "client connection timeout")
	flag.Parse()

	openedConnections := sync.WaitGroup{}
	openedConnections.Add(1)

	var server *filet.Server
	var client *filet.Client
	if *runServer {
		server = &filet.Server{
			Address: &utils.Address{
				Proto: *proto,
				Addr:  *address,
				Port:  port,
			},
			Clients:          make([]*net.Conn, 5),
			ConnectionWaiter: &openedConnections,
			RequestHandler: func(client *net.Conn, received *requests.Request) {

				fmt.Printf("Received packet %v\n", (*received).Info().Id)

				if (*received).Info().WantsResponse {
					_, _, _ = utils.SendRequestOn(client, (*received).GetResult())
				}
			},
		}
		go server.Start()
		defer fio.ClearDir("dl")
		defer server.Close()

	} else {

		client = (&filet.Client{
			Address: &utils.Address{
				Proto: *proto,
				Addr:  *address,
				Port:  port,
			},
		}).Start(*timeout)
		defer client.Close()

		client.Send(defaultRequests.MakeTextRequest("test123 test 1 2 1 2 test 1 2 3"))
		client.Send(defaultRequests.MakeFileRequest("./res/test.txt", false))
		client.Send(defaultRequests.MakeTextRequest("à Kadoc"))
		client.Send(defaultRequests.MakeFileRequest("./res/8.png", true))
		client.Send(defaultRequests.MakeTextRequest("en garde ma mignonne"))
	}

	terminalInput(server, &openedConnections)
	openedConnections.Wait()
}

func terminalInput(server *filet.Server, group *sync.WaitGroup) {

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		line := scanner.Text()
		fmt.Printf("got > %v\n", line)
		if line == "stop" {
			if server != nil {
				server.Close()
			}
			group.Done()
			break
		}
	}
}
