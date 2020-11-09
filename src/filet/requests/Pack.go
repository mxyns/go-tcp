package requests

import (
	"encoding/binary"
	"fmt"
	"net"
)

type Pack struct {
	dataCount  uint32
	requests   []*Request
	packedInfo *RequestInfo
}

// No need to register, it's registered by default using previously Reserved Id 0

func MakePack(requests ...Request) *Pack {

	packResponse := false
	requestsPointers := make([]*Request, len(requests))
	for i := range requests {
		packResponse = packResponse || (requests[i]).Info().WantsResponse
		requestsPointers[i] = &requests[i]
	}

	return &Pack{
		packedInfo: &RequestInfo{
			Id:            0,
			WantsResponse: packResponse,
		},
		dataCount: uint32(len(requests)),
		requests:  requestsPointers,
	}
}

func (p *Pack) SerializeTo(conn *net.Conn) error {

	var err error
	for i := range p.requests {

		err, _ = WriteHeaderTo(conn, p.requests[i])
		if err != nil {
			return err
		}

		err = (*p.requests[i]).SerializeTo(conn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Pack) DeserializeFrom(conn *net.Conn) (Request, error) {

	// load Pack size
	length := make([]byte, 32/8)
	_, err := (*conn).Read(length)
	if err != nil {
		return p, err
	}
	p.dataCount = binary.BigEndian.Uint32(length)

	// capture the following p.dataCount requests and put them in Pack
	fmt.Printf("Receiving pack of %v Requests\n", p.dataCount)
	p.requests = make([]*Request, p.dataCount)
	for i := range p.requests {
		req, err, _ := Await(conn)
		if err != nil {
			return p, err
		}
		p.requests[i] = req
	}

	return p, nil
}

func (p *Pack) Name() string       { return "Pack" }
func (p *Pack) Info() *RequestInfo { return p.packedInfo }
func (p *Pack) DataSize() uint32   { return uint32(len(p.requests)) }

func (p *Pack) GetResult() Request {

	results := make([]Request, len(p.requests))
	for i := range p.requests {
		result := (*p.requests[i]).GetResult()
		if result == nil {
			result = MakePack()
		}
		results[i] = result
	}

	return MakePack(results...)
}

func (p *Pack) GetCount() uint32 {
	return p.dataCount
}
func (p *Pack) GetRequests() []*Request {
	return p.requests
}
