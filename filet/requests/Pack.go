package requests

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
)

const (
	HalfUInt32 = math.MaxUint32 / 2
)

type Pack struct {
	dataCount  uint32
	requests   []*Request
	packedInfo *RequestInfo

	/* indicative Id for receiver
	from 0 -> max(uint32)/2 Id for Request
	from max(uint32)/2 + 1 -> max(uint32) Id for Response / Result
	Result Id = Request Id + max(uint16)
		=> 	we know when receiving a Request if it's a reply or not
			by : isReply = subId > max(uint32) / 2
	*/

	SubId   uint32
	IsReply bool
}

// No need to register, it's registered by default using previously Reserved Id 0

func MakeGenericPack(requests ...Request) *Pack {
	return MakePack(0, requests...)
}
func MakePack(subId uint32, requests ...Request) *Pack {

	packResponse := false
	requestsPointers := make([]*Request, len(requests))
	for i := range requests {
		packResponse = packResponse || (requests[i]).Info().WantsResponse
		requestsPointers[i] = &requests[i]
	}

	return &Pack{
		SubId: subId,
		packedInfo: &RequestInfo{
			Id:            0,
			WantsResponse: packResponse,
		},
		dataCount: uint32(len(requests)),
		requests:  requestsPointers,
	}
}

func (p *Pack) SerializeTo(conn *net.Conn) error {

	err := binary.Write(*conn, binary.BigEndian, p.SubId)
	if err != nil {
		fmt.Printf("Error while sending SubId : %v\n", err)
		return err
	}

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

	// load subId
	subId := make([]byte, 32/8)
	_, err = (*conn).Read(subId)
	if err != nil {
		return p, err
	}
	p.SubId = binary.BigEndian.Uint32(subId)

	p.IsReply = p.SubId > HalfUInt32
	if p.IsReply {
		p.SubId -= HalfUInt32 + 1
	}

	// capture the following p.dataCount requests and put them in Pack
	fmt.Printf("Receiving (reply=%v) pack sub %v of %v Requests\n", p.IsReply, p.SubId, p.dataCount)
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
			result = MakeGenericPack()
			result.(*Pack).SubId = HalfUInt32 + 1
		}
		results[i] = result
	}

	return MakePack(1+HalfUInt32+p.SubId, results...)
}

func (p *Pack) GetCount() uint32 {
	return p.dataCount
}
func (p *Pack) GetRequests() []*Request {
	return p.requests
}
