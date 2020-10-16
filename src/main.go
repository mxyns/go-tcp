package main

import (
	"bufio"
	"fil"
	"fil/endpoints"
	"fil/requests"
	fio "fileio"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

var runServer, _ = strconv.ParseBool(os.Args[1:][0])

func main() {

	openedConnections := sync.WaitGroup {}
	openedConnections.Add(1)

	var server *fil.Server
	var client *fil.Client
	if runServer {
		server = &fil.Server{
			Address: &endpoints.Address{
				Proto: "tcp",
				Addr:  "",
				Port:  8888,
			},
			Clients:          make([]*net.Conn, 5),
			ConnectionWaiter: &openedConnections,
		}
		go server.Start()
		defer fio.ClearDir("dl")

	} else {

		client = &fil.Client{
			Address: &endpoints.Address{
				Proto: "tcp",
				Addr:  "127.0.0.1",
				Port:  8888,
			},
		}
		client.Start()
		defer client.Close()

		client.Send(requests.MakeTextRequest("test123 test 1 2 1 2 test 1 2 3"))
		client.Send(requests.MakeFileRequest("./test.txt"))
		client.Send(requests.MakeTextRequest("à Kadoc"))
		client.Send(requests.MakeFileRequest("./8.png"))
		client.Send(requests.MakeTextRequest("en garde ma mignonne"))
	}

	terminalInput(client, server, &openedConnections)
	openedConnections.Wait()
}

func terminalInput(client *fil.Client, server *fil.Server, group *sync.WaitGroup) {

	scanner := bufio.NewScanner(os.Stdin)
	for {
		scanner.Scan()
		line := scanner.Text()
		if line == "stop" {
			fmt.Printf("got > %v\n", line)
			if runServer {
				server.Close()
				group.Done()
				break
			} else {
				client.Close()
				break
			}
		}
	}
}