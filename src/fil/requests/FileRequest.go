package requests

import (
	"encoding/binary"
	fio "fileio"
	"fmt"
	"net"
	"os"
	"strings"
)

type fileRequest struct {
	info          *RequestInfo
	in_path       string
	out_path      string
	wantsResponse bool
}

func init() {

	RegisterRequestType(2, func(reqInfo *RequestInfo) Request { return &fileRequest{info: reqInfo} })
}

func MakeFileRequest(in string, wantsResponse bool) *fileRequest {

	return &fileRequest{
		info:    &RequestInfo{Id: 2, WantsResponse: wantsResponse},
		in_path: in,
	}
}

func (fr *fileRequest) Name() string       { return "File" }
func (fr *fileRequest) Info() *RequestInfo { return fr.info }
func (fr *fileRequest) DataSize() uint32 {
	stat, err := os.Stat(fr.in_path)
	if err != nil || stat.IsDir() || stat.Size() > 1<<32-1 {
		fmt.Printf("Error while evaluating file %v's size : %v\n", fr.in_path, err)
		return 0
	}

	return uint32(stat.Size())
}

func (fr *fileRequest) SerializeTo(conn *net.Conn) {

	_, err := fio.StreamFromFile(conn, fr.in_path)
	if err != nil {
		// FIXME fill the gap with zeros ?
		fmt.Printf("Error while serializing file. Can't continue. Closing connection.\n%v", err)
		(*conn).Close()
	}
}
func (fr *fileRequest) DeserializeFrom(conn *net.Conn) Request {

	length := make([]byte, 32/8)
	_, _ = (*conn).Read(length)
	data_length := binary.BigEndian.Uint32(length)

	fmt.Printf("Received length : %v\n", data_length)
	name, read, err := fio.WriteStreamToFile(conn, "dl", data_length)

	if err == nil {
		fr.out_path = name
		fmt.Printf("File received, %v bytes saved to %v.\n", read, fr.out_path)
		return fr
	}

	return nil
}

func (fr *fileRequest) GetResult() Request {

	shards := strings.Split(fr.out_path, "/")

	// TODO make it a File Id Response/Request kinda thing
	return MakeTextRequest(shards[len(shards)-1])
}
