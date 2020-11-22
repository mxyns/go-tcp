package fileio

import (
	"bufio"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"net"
	"os"
	"strconv"
	"time"
)

const MAX_FILE_BUFFER_SIZE = 8192

func ClearDir(path string) {

	path = "./" + path
	err := os.RemoveAll(path)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
			"path":  path,
		}).Error("Error while clearing directory")
	}
	_ = os.Mkdir(path, os.ModeDir)
}

func StreamFromFile(conn *net.Conn, file_path string) (uint32, error) {

	f, err := os.Open(file_path)
	if err != nil {
		return 0, err
	}
	r := bufio.NewReader(f)
	b := make([]byte, MAX_FILE_BUFFER_SIZE)

	sum := uint32(0)
	for {
		n_read, err := r.Read(b)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Info("Stopped reading")
			if err == io.EOF {
				return sum, nil
			} else {
				return sum, err
			}
		}
		n_sent, err := (*conn).Write(b[0:n_read])
		if err != nil || n_sent != n_read {
			log.WithFields(log.Fields{
				"error": err,
				"read":  n_read,
				"sent":  n_sent,
			}).Error("Didn't send all info I read")
			return sum, err
		}
		sum += uint32(n_sent)
	}
}

func WriteStreamToFile(conn *net.Conn, folder string, data_length uint32) (string, uint32, error) {

	var buffer []byte
	buffer = make([]byte, MAX_FILE_BUFFER_SIZE)

	f, err := os.Create("./" + folder + "/" + strconv.Itoa(time.Now().Nanosecond()+rand.Int()))
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Couldn't create file. Aborting")
		defer log.WithFields(log.Fields{
			"result": (*conn).Close(),
		}).Error("Closing connection to Client")
		return "", 0, err
	}

	writer := bufio.NewWriter(f)
	defer f.Close()

	read := uint32(0)
	for {
		if left := data_length - read; left < MAX_FILE_BUFFER_SIZE {
			buffer = make([]byte, left)
		}

		n_read, err := (*conn).Read(buffer)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Stopped reading")
			break
		} else {
			log.WithFields(log.Fields{
				"count":    n_read,
				"capacity": len(buffer),
			}).Info("read bytes from buffer")
		}
		n_saved, err := writer.Write(buffer[0:n_read])
		writer.Flush()

		if err != nil || n_saved != n_read {
			log.WithFields(log.Fields{
				"error": err,
				"read":  n_read,
				"saved": n_saved,
			}).Error("Didn't save all info I read")
			break
		}
		log.Info("%v written \n", n_saved)
		log.WithFields(log.Fields{
			"count": n_saved,
		}).Error("wrote bytes")

		if read += uint32(n_saved); read >= data_length {
			return f.Name(), read, nil
		}
	}

	return "", 0, err
}
