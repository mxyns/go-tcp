package requests

import (
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
)

const (
	MAX_PACK_SUBID = math.MaxUint32 / 2
	PACK_ID        = 0
)

type Pack struct {
	dataCount  uint32
	requests   []*Request
	packedInfo *RequestInfo

	/*
		indicative Id for receiver
		from 0 -> max(uint32)/2 Id for Request
		from max(uint32)/2 + 1 -> max(uint32) Id for Response / Result
		Result Id = Request Id + max(uint16)
			=> 	we know when receiving a Request if it's a reply or not
				by : isReply = subId > max(uint32) / 2
	*/

	SubId   uint32
	IsReply bool
}

// No need to register, it's registered by defaultRequests using (previously reserved) Id 0

func MakeGenericPack(requests ...Request) *Pack {
	return MakePack(PACK_ID, requests...)
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
			Id:            PACK_ID,
			WantsResponse: packResponse,
		},
		dataCount: uint32(len(requests)),
		requests:  requestsPointers,
	}
}

func (p *Pack) SerializeTo(conn *net.Conn) error {

	err := binary.Write(*conn, binary.BigEndian, p.SubId)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while sending SubId")
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

	p.IsReply = p.SubId > MAX_PACK_SUBID
	if p.IsReply {
		p.SubId -= MAX_PACK_SUBID + 1
	}

	// capture the following p.dataCount requests and put them in Pack
	log.WithFields(log.Fields{
		"subId": p.SubId,
		"count": p.dataCount,
		"reply": p.IsReply,
	}).Info("Receiving pack requests")
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
			result.(*Pack).SubId = MAX_PACK_SUBID + 1
		}
		results[i] = result
	}

	return MakePack(1+MAX_PACK_SUBID+p.SubId, results...)
}

func (p *Pack) GetCount() uint32 {
	return p.dataCount
}
func (p *Pack) GetRequests() []*Request {
	return p.requests
}
