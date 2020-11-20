package defaultRequests

import (
	"encoding/binary"
	"fmt"
	fio "github.com/mxyns/go-tcp/fileio"
	"github.com/mxyns/go-tcp/filet/requests"
	"net"
	"os"
	"strings"
)

const (
	folderName = "dl"
)

type FileRequest struct {
	info          *requests.RequestInfo
	path          string // path of file (input path if on sender's side, output path on receiver's side)
	filesize      uint32
	wantsResponse bool
}

func init() {

	requests.RegisterRequestType(2, func(reqInfo *requests.RequestInfo) requests.Request { return &FileRequest{info: reqInfo} })
	if _, err := os.Stat(folderName); os.IsNotExist(err) {
		_ = os.MkdirAll(folderName, os.ModeDir)
	}
}

func MakeFileRequest(in string, wantsResponse bool) *FileRequest {

	fileRequest := &FileRequest{
		info: &requests.RequestInfo{Id: 2, WantsResponse: wantsResponse},
		path: in,
	}
	fileRequest.filesize = fileRequest.DataSize()

	return fileRequest
}

func (fr *FileRequest) Name() string                { return "File" }
func (fr *FileRequest) Info() *requests.RequestInfo { return fr.info }
func (fr *FileRequest) DataSize() uint32 {
	stat, err := os.Stat(fr.path)
	if err != nil || stat.IsDir() || stat.Size() > 1<<32-1 {
		fmt.Printf("Error while evaluating file %v's size : %v\n", fr.path, err)
		return 0
	}

	return uint32(stat.Size())
}

func (fr *FileRequest) SerializeTo(conn *net.Conn) error {

	_, err := fio.StreamFromFile(conn, fr.path)
	if err != nil {
		// TODO fill the gap with zeros ?
		return err
	}

	return err
}
func (fr *FileRequest) DeserializeFrom(conn *net.Conn) (requests.Request, error) {

	length := make([]byte, 32/8)
	_, _ = (*conn).Read(length)
	data_length := binary.BigEndian.Uint32(length)
	name, wrote, err := fio.WriteStreamToFile(conn, folderName, data_length)

	if err != nil {
		return nil, err
	}

	fr.path = name
	fr.filesize = wrote
	fmt.Printf("File received, %v bytes saved to %v.\n", wrote, fr.path)
	return fr, nil
}

func (fr *FileRequest) GetResult() requests.Request {

	shards := strings.Split(fr.path, "/")

	return MakeTextRequest(shards[len(shards)-1])
}

func (fr *FileRequest) GetPath() string {
	return fr.path
}
func (fr *FileRequest) GetFileSize() uint32 {
	return fr.filesize
}
