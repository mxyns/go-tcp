package requests

import (
	"encoding/binary"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io"
	"net"
)

/*
	Error id : 0 => fine
        [0; 255] => wrong id received
      [255; 512] => io error (see consts)
*/
const (
	CONN_WRITE_ERROR uint16 = 256
	CONN_READ_ERROR  uint16 = 257
)

func WriteHeaderTo(conn *net.Conn, request *Request) (err error, err_id uint16) {

	err = binary.Write(*conn, binary.BigEndian, (*request).Info().Id)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while sending Id")
		return err, CONN_WRITE_ERROR
	}

	err = binary.Write(*conn, binary.BigEndian, (*request).Info().WantsResponse)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while sending NeedsReponse")
		return err, CONN_WRITE_ERROR
	}

	err = binary.Write(*conn, binary.BigEndian, (*request).DataSize())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while sending DataSize")
		return err, CONN_WRITE_ERROR
	}

	return nil, 0
}

func SendRequestOn(conn *net.Conn, request *Request) (req *Request, err error, err_id uint16) {

	log.WithFields(log.Fields{
		"name": (*request).Name(),
		"id":   (*request).Info().Id,
	}).Debug("Sending request")

	err, err_id = WriteHeaderTo(conn, request)

	err = (*request).SerializeTo(conn)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error while serializing request")
		return nil, err, CONN_WRITE_ERROR
	}

	if (*request).Info().WantsResponse {
		req, err, err_id = Await(conn)
		return req, err, err_id
	}

	return nil, nil, 0
}
func Await(socket *net.Conn) (req *Request, err error, err_id uint16) {

	twoByteBuff := make([]byte, 2)

	n, err := io.ReadFull(*socket, twoByteBuff)
	if err != nil {
		return nil, err, CONN_READ_ERROR
	}

	id := twoByteBuff[0]

	wantsResponse := twoByteBuff[1] == 1

	log.WithFields(log.Fields{
		"byteCount":     n,
		"id":            id,
		"wantsResponse": wantsResponse,
	}).Info("Read header")
	log.WithFields(log.Fields{
		"id":      id,
		"address": (*socket).RemoteAddr(),
	}).Debug("Received request client")

	received := (&RequestInfo{Id: id, WantsResponse: wantsResponse}).BuildFrom(socket)
	if received == nil {
		return nil, fmt.Errorf("unknown Id"), uint16(id)
	}

	return &received, nil, 0
}
