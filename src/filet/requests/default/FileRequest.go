package _default

import (
	"encoding/binary"
	fio "fileio"
	"filet/requests"
	"fmt"
	"net"
	"os"
	"strings"
)

type fileRequest struct {
	info          *requests.RequestInfo
	in_path       string
	out_path      string
	wantsResponse bool
}

func init() {

	requests.RegisterRequestType(2, func(reqInfo *requests.RequestInfo) requests.Request { return &fileRequest{info: reqInfo} })
}

func MakeFileRequest(in string, wantsResponse bool) *fileRequest {

	return &fileRequest{
		info:    &requests.RequestInfo{Id: 2, WantsResponse: wantsResponse},
		in_path: in,
	}
}

func (fr *fileRequest) Name() string                { return "File" }
func (fr *fileRequest) Info() *requests.RequestInfo { return fr.info }
func (fr *fileRequest) DataSize() uint32 {
	stat, err := os.Stat(fr.in_path)
	if err != nil || stat.IsDir() || stat.Size() > 1<<32-1 {
		fmt.Printf("Error while evaluating file %v's size : %v\n", fr.in_path, err)
		return 0
	}

	return uint32(stat.Size())
}

func (fr *fileRequest) SerializeTo(conn *net.Conn) error {

	_, err := fio.StreamFromFile(conn, fr.in_path)
	if err != nil {
		// FIXME fill the gap with zeros ?
		return err
	}

	return err
}
func (fr *fileRequest) DeserializeFrom(conn *net.Conn) (requests.Request, error) {

	length := make([]byte, 32/8)
	_, _ = (*conn).Read(length)
	data_length := binary.BigEndian.Uint32(length)
	name, read, err := fio.WriteStreamToFile(conn, "dl", data_length)

	if err != nil {
		return nil, err
	}

	fr.out_path = name
	fmt.Printf("File received, %v bytes saved to %v.\n", read, fr.out_path)
	return fr, nil
}

func (fr *fileRequest) GetResult() requests.Request {

	shards := strings.Split(fr.out_path, "/")

	return MakeTextRequest(shards[len(shards)-1])
}
