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

	requestRegister[0] = func(packInfo *RequestInfo) Request { return &Pack{packedInfo: packInfo} } // Réservé
}

// Element, Primitive, Shard, Extract,
type Request interface {
	SerializeTo(conn *net.Conn) error
	DeserializeFrom(conn *net.Conn) (Request, error)
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

		received, err := requestRegister[req.Id](req).DeserializeFrom(conn)
		if err != nil {
			fmt.Printf("Error while deserializing request. Stopping. %v", err)
			(*conn).Close()
			return nil
		}
		return received

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
