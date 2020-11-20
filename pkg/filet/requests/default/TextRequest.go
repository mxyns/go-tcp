package _default

import (
	"encoding/binary"
	"fmt"
	"io"
	"mxyns/go-tcp/filet/requests"
	"net"
)

type TextRequest struct {
	info *requests.RequestInfo
	text string
}

func init() {
	requests.RegisterRequestType(1, func(reqInfo *requests.RequestInfo) requests.Request { return &TextRequest{info: reqInfo} })
}

func MakeTextRequest(text string) *TextRequest {
	return &TextRequest{
		info: &requests.RequestInfo{Id: 1, WantsResponse: false},
		text: text,
	}
}

func (tr *TextRequest) Name() string                { return "Text" }
func (tr *TextRequest) Info() *requests.RequestInfo { return tr.info }
func (tr *TextRequest) DataSize() uint32            { return uint32(len([]byte(tr.text))) }

func (tr *TextRequest) SerializeTo(conn *net.Conn) error {

	data := []byte(tr.text)
	n, err := (*conn).Write(data)
	if n != len(data) {
		return fmt.Errorf("didn't send as much text as I had : %v\n", err)
	}

	return nil
}
func (tr *TextRequest) DeserializeFrom(conn *net.Conn) (requests.Request, error) {

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

func (tr *TextRequest) GetResult() requests.Request {

	return nil
}

func (tr *TextRequest) GetText() string {
	return tr.text
}
