package endpoints

import (
	"encoding/binary"
	"fil/requests"
	"fmt"
	"io"
	"net"
)

func SendRequestOn(conn *net.Conn, request requests.Request) *requests.Request {

	fmt.Printf("Sending Request [%v]\n", request.Name())

	fmt.Printf("Id : %v\n",request.Info().Id)
	n, err := (*conn).Write([]byte{request.Info().Id})
	if n != 1 || err != nil {
		fmt.Printf("Error while sending Id : %v\n", err)
		return nil
	}

	err = binary.Write(*conn, binary.BigEndian, request.DataSize())
	if err != nil {
		fmt.Printf("Error while sending DataSize : %v\n", err)
		return nil
	}

	request.SerializeTo(conn)

	if request.NeedsResponse() {
		return Await(conn)
	}

	return nil
}
func Await(socket *net.Conn) *requests.Request {

	oneByteBuff := make([]byte, 1)

	n, err := io.ReadFull(*socket, oneByteBuff)
	if err != nil {
		fmt.Printf("Couldn't read from socket. Connection must be closed. %v\n", err)
		return nil
	}

	id := oneByteBuff[0]
	fmt.Printf("Read %v byte for Id=%v\n", n, id)
	fmt.Printf("Received Request[%v] from %v\n", id, (*socket).RemoteAddr())

	received := (&requests.RequestInfo { Id: id }).BuildFrom(socket)

	return &received
}