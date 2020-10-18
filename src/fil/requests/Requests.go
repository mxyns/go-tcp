package requests

import (
	"fmt"
	"net"
)

/*
   id == 0 => reservé
      == 1 => image
  		=> length => read => save => send id
      == 1 => filtres + image
		=> filter count => filter ids
			=> length (image) => read => apply => send result
*/

var requestRegister = make(map[byte]func(*RequestInfo) Request)

func init() {

	requestRegister[0] = nil // Réservé
}

type Request interface {
	SerializeTo(conn *net.Conn)
	DeserializeFrom(conn *net.Conn) Request
	DataSize() uint32
	Name() string
	Info() *RequestInfo
	GetResult() Request
}
type RequestInfo struct {
	Id            byte
	WantsResponse bool
}

func (req *RequestInfo) BuildFrom(conn *net.Conn) Request {

	tyqe := requestRegister[req.Id]
	if tyqe != nil {
		return requestRegister[req.Id](req).DeserializeFrom(conn)
	} else {
		return nil
	}
}
func RegisterRequestType(id byte, generator func(reqInfo *RequestInfo) Request) {

	tyqe := generator(&RequestInfo{Id: 0}).Name()

	if requestRegister[id] == nil {
		fmt.Printf("Registered %v as Id=%v\n", tyqe, id)
		requestRegister[id] = generator
	} else {
		used_type := requestRegister[id](&RequestInfo{Id: 0}).Name()
		panic("Failed to register type " + tyqe + ". This Id is already in use by Request Type " + used_type)
	}
}
