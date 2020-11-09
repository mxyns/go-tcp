package main

import (
	"bufio"
	fio "fileio"
	"filet"
	"filet/requests"
	defaultRequests "filet/requests/default"
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
	if *runServer {
		server = &filet.Server{
			Address: &filet.Address{
				Proto: *proto,
				Addr:  *address,
				Port:  port,
			},
			Clients:          make([]*net.Conn, 5),
			ConnectionWaiter: &openedConnections,
			RequestHandler: func(client *net.Conn, received *requests.Request) {

				printRequest(received, fmt.Sprintf("%v ==> ", (*client).RemoteAddr()))

				if (*received).Info().WantsResponse {
					response := (*received).GetResult()
					_, _, _ = requests.SendRequestOn(client, &response)
					printRequest(&response, fmt.Sprintf("%v <== ", (*client).RemoteAddr()))
				}
			},
		}
		go server.Start()
		defer fio.ClearDir("dl")
		defer server.Close()

	} else {

		client, err := (&filet.Client{
			Address: &filet.Address{
				Proto: *proto,
				Addr:  *address,
				Port:  port,
			},
		}).Start(*timeout)

		if err != nil {
			fmt.Printf("Client couldn't connect to server. %v", err)
			return
		}
		defer client.Close()

		client.Send(defaultRequests.MakeTextRequest("test123 test 1 2 1 2 test 1 2 3"))
		client.Send(defaultRequests.MakeFileRequest("./examples/demo/res/test.txt", false))
		client.Send(defaultRequests.MakeTextRequest("à Kadoc"))
		client.Send(requests.MakePack(
			defaultRequests.MakeTextRequest("Un peu de texte avec deux pièces-jointes (à deux doigts d'inventer le mail)"),
			defaultRequests.MakeFileRequest("./examples/demo/res/test.txt", true),
			defaultRequests.MakeFileRequest("./examples/demo/res/test.txt", true),
		))
		client.Send(defaultRequests.MakeFileRequest("./examples/demo/res/jc.jpg", true))
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

func printRequest(request *requests.Request, prefix string) {
	if request == nil {
		return
	} // when pack results are nil

	switch (*request).(type) {
	case *requests.Pack: // pack
		pack := (*request).(*requests.Pack)
		fmt.Printf("%vPack of %v requests : %v\n", prefix, pack.GetCount(), pack.GetRequests())
		for i := range pack.GetRequests() {
			printRequest(pack.GetRequests()[i], fmt.Sprintf("    %v. ", i))
		}

	case *defaultRequests.TextRequest: // textRequest
		fmt.Printf("%v%v\n", prefix, (*request).(*defaultRequests.TextRequest).GetText())
	case *defaultRequests.FileRequest: // fileRequest
		fmt.Printf("%vFile : path=%v, size=%v\n", prefix, (*request).(*defaultRequests.FileRequest).GetPath(), (*request).(*defaultRequests.FileRequest).GetFileSize())
	default:
		fmt.Printf("%vReceived packet %v\n", prefix, (*request).Info().Id)
	}
}
