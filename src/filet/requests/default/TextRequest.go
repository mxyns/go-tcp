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

func (tr *textRequest) SerializeTo(conn *net.Conn) {

	data := []byte(tr.text)
	n, err := (*conn).Write(data)
	if n != len(data) {
		fmt.Printf("Didn't send as much text as I had : %v\n", err)
	}
}
func (tr *textRequest) DeserializeFrom(conn *net.Conn) requests.Request {

	length := make([]byte, 32/8)
	_, _ = (*conn).Read(length)
	fmt.Printf("Received length : %v\n", binary.BigEndian.Uint32(length))
	data := make([]byte, binary.BigEndian.Uint32(length))
	n, _ := io.ReadFull(*conn, data)
	tr.text = string(data)
	fmt.Printf("Read %v bytes giving me : '%v'\n", n, tr.text)

	return tr
}

func (tr *textRequest) GetResult() requests.Request {

	return nil
}