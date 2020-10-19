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
```go
import (
    fil "github.com/mxyns/go-tcp/src/filet" // [compulsory] main package (Client, Server, Address, Communication functions)
    req "github.com/mxyns/go-tcp/src/filet/requests" // [compulsory] request support 
    defReq "github.com/mxyns/go-tcp/src/filet/requests/default" // [optional] some default requests 
)
```

# using the package
## server
example took from [demo.go](https://github.com/mxyns/go-tcp/blob/master/example/demo.go)
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
			response := (*received).GetResult()
			_, _, _ = utils.SendRequestOn(client, &response)
	        }
        },
    }
go server.Start() // Start Server in goroutine
defer server.Close() // Close server when finished

openedConnections.Wait()
```

## client
example took from [demo.go](https://github.com/mxyns/go-tcp/blob/master/example/demo.go)
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
	
if err != nil { // Check if the Client connected correctly
	fmt.Printf("Client couldn't connect to server. %v", err)
	return
}
defer client.Close() // Close Client when finished

client.Send(defaultRequests.MakeTextRequest("Connected !")) // Send text using default request from filet/requests/default package
client.Send(defaultRequests.MakeFileRequest("./path/to/file.ext")) // Send file using default request from filet/requests/default package

waitForSomething.Wait()
```

# making a custom request
check implementations in the [default requests package](https://github.com/mxyns/go-tcp/tree/master/src/filet/requests/default)
or use the following template : 
```go
package some_package

import (
	"encoding/binary"
	"filet/requests"
	"net"
)

type customRequest struct {
	info            *requests.RequestInfo
	someTrait       interface{}
}

func init() {

    // Register request type, if not done the request can't be treated 
    requests.RegisterRequestType(ID, func(reqInfo *requests.RequestInfo) requests.Request { return &customRequest{info: reqInfo} })
}

// Used to create your type of requests
func MakeCustomRequest(my_parameters interface{}, wantsResponse bool) *customRequest {

	return &customRequest{
		info:    &requests.RequestInfo{Id: ID, WantsResponse: wantsResponse},
		someTrait: my_parameters,
	}
}

// Name is only used for debug / display
func (cr *customRequest) Name() string                { return "My Name" }

// DataSize is the length in bytes of the content put by SerializeTo. It's put in the request header when it's sent 
func (cr *customRequest) DataSize() uint32 { return uint32(len(cr.someTrait)) }
func (cr *customRequest) Info() *requests.RequestInfo { return cr.info }

// Serializes and writes data to the socket
func (cr *customRequest) SerializeTo(conn *net.Conn) error {

    _, err := (*conn).Write(cr.someTrait.toByteArray())
    return err
}

// Reads data from the socket and fills the struct's fields 
func (cr *customRequest) DeserializeFrom(conn *net.Conn) (requests.Request, error) {

    length := make([]byte, 32/8)
    _, err := (*conn).Read(length)
    if err != nil { return nil, err }
    data_length := binary.BigEndian.Uint32(length)
	
    data := make([]byte, data_length)
    _, err = (*conn).Read(data)
    if err != nil { return nil, err }
    cr.someTrait = someTrait{}.From(data)

    return cr, nil
}

// The response that should be sent after reception
func (cr *customRequest) GetResult() requests.Request {

	return MakeACertainRequest(some_parameters)
}
```