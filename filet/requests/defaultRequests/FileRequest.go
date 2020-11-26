package defaultRequests

import (
	"encoding/binary"
	fio "github.com/mxyns/go-tcp/fileio"
	"github.com/mxyns/go-tcp/filet/requests"
	log "github.com/sirupsen/logrus"
	"math"
	"net"
	"os"
	"strings"
)

var (
	TARGET_DIRECTORY = "dl"
)

const (
	FILE_REQUEST_ID = 2
)

type FileRequest struct {
	info          *requests.RequestInfo
	path          string // path of file (input path if on sender's side, output path on receiver's side)
	filesize      uint32
	wantsResponse bool
}

func init() {

	requests.RegisterRequestType(FILE_REQUEST_ID, func(reqInfo *requests.RequestInfo) requests.Request { return &FileRequest{info: reqInfo} })
	if _, err := os.Stat(TARGET_DIRECTORY); os.IsNotExist(err) {
		_ = os.MkdirAll(TARGET_DIRECTORY, os.ModeDir)
	}
}

func MakeFileRequest(in string, wantsResponse bool) *FileRequest {

	fileRequest := &FileRequest{
		info: &requests.RequestInfo{Id: FILE_REQUEST_ID, WantsResponse: wantsResponse},
		path: in,
	}
	fileRequest.filesize = fileRequest.DataSize()

	return fileRequest
}

func (fr *FileRequest) Name() string                { return "File" }
func (fr *FileRequest) Info() *requests.RequestInfo { return fr.info }
func (fr *FileRequest) DataSize() uint32 {
	stat, err := os.Stat(fr.path)
	if err != nil || stat.IsDir() || stat.Size() > math.MaxUint32 {
		log.WithFields(log.Fields{
			"path":  fr.path,
			"error": err,
		}).Error("Error while evaluating file size")
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

	length := make([]byte, requests.LENGTH_SIZE)
	_, _ = (*conn).Read(length)
	data_length := binary.BigEndian.Uint32(length)
	name, wrote, err := fio.WriteStreamToFile(conn, TARGET_DIRECTORY, data_length)

	if err != nil {
		return nil, err
	}

	fr.path = name
	fr.filesize = wrote
	log.WithFields(log.Fields{
		"path": fr.path,
		"size": wrote,
	}).Info("File received.")
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
