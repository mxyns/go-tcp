# go-tcp
TCP Server / Client pair with customizable requests types

# running demo
launch arguments :
* `-s true` server mode (default mode is client / false)
* `-P tcp` protocol (idk if it works with something else)
* `-p 888` port
* `-a 0.0.0.0` address to connect to / open server on
* `-t 1000` connection timeout (client) in milliseconds

# importing the package
`import "github.com/mxyns/go-tcp/"`

# using the package
## server
```go
openedConnections := sync.WaitGroup{}
openedConnections.Add(1)

// Make a server
server := &filet.Server{
			Address: &utils.Address{
				Proto: "tcp", // Server protocol 
				Addr:  "", // Open on all adapters 
				Port:  8888, // Listen on port 8888 
			},
			Clients:          make([]*net.Conn, 5), // Default clients slice
			ConnectionWaiter: &openedConnections, // sync.WaitGroup type. Counts connected clients 
			RequestHandler: func(client *net.Conn, received *requests.Request) { // Function called on Request reception

				fmt.Printf("Received packet %v\n", (*received).Info().Id)

				if (*received).Info().WantsResponse {
					_, _, _ = utils.SendRequestOn(client, (*received).GetResult())
				}
			},
		}
		go server.Start() // Start Server in goroutine
		defer server.Close() // Close server when finished

	    openedConnections.Wait()
```

## client
```go
waitForSomething := sync.WaitGroup{}
waitForSomething.Add(1)

client := (&filet.Client{
			Address: &utils.Address{
				Proto: "tcp", // Protocol to use
				Addr:  "127.0.0.1", // Server IP to connect to
				Port:  8888, // Port the Server is listening on
			},
		}).Start("10s") // Try to connect and timeout after 10 sec
		defer client.Close() // Close Client when finished

client.Send(defaultRequests.MakeTextRequest("Connected !")) // Send default request from filet/requests/default package
waitForSomething.Wait()
```
