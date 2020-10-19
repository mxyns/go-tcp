package _default

import (
	"encoding/binary"
	"filet/requests"
	"fmt"
	"io"
	"net"
)

type textRequest struct {
	info *requests.RequestInfo
	text string
}

func init() {
	requests.RegisterRequestType(1, func(reqInfo *requests.RequestInfo) requests.Request { return &textRequest{info: reqInfo} })
}

func MakeTextRequest(text string) *textRequest {
	return &textRequest{
		info: &requests.RequestInfo{Id: 1, WantsResponse: false},
		text: text,
	}
}

func (tr *textRequest) Name() string                { return "Text" }
func (tr *textRequest) Info() *requests.RequestInfo { return tr.info }
func (tr *textRequest) DataSize() uint32            { return uint32(len([]byte(tr.text))) }

func (tr *textRequest) SerializeTo(conn *net.Conn) error {

	data := []byte(tr.text)
	n, err := (*conn).Write(data)
	if n != len(data) {
		return fmt.Errorf("didn't send as much text as I had : %v\n", err)
	}

	return nil
}
func (tr *textRequest) DeserializeFrom(conn *net.Conn) (requests.Request, error) {

	length := make([]byte, 32/8)
	_, err := (*conn).Read(length)
	if err != nil {
		return tr, err
	}
	data := make([]byte, binary.BigEndian.Uint32(length))
	_, err = io.ReadFull(*conn, data)

	if err != nil {
		return tr, err
	}

	tr.text = string(data)

	return tr, err
}

func (tr *textRequest) GetResult() requests.Request {

	return nil
}
