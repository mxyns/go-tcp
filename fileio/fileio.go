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

	err := os.RemoveAll("./" + path)
	if err != nil {
		log.Error("Error while clearing directory ./%v", path)
	}
	_ = os.Mkdir("./"+path, os.ModeDir)
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
			log.Info("Stopped reading : %v\n", err)
			if err == io.EOF {
				return sum, nil
			} else {
				return sum, err
			}
		}
		n_sent, err := (*conn).Write(b[0:n_read])
		if err != nil || n_sent != n_read {
			log.Error("Didn't send all info I read : %v\n", err)
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
		log.Error("Couldn't create file. Aborting.%v\n", err)
		defer log.Error("Closing connection to Client : %v\n", (*conn).Close())
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
			log.Error("Stopped reading : %v\n", err)
			break
		} else {
			log.Info("%v bytes out of %v buffer capacity\n", n_read, len(buffer))
		}
		n_saved, err := writer.Write(buffer[0:n_read])
		writer.Flush()

		if err != nil || n_saved != n_read {
			log.Error("Didn't save all info I read %v != %v : %v\n%", n_read, n_saved, err)
			break
		}
		log.Info("%v written \n", n_saved)

		if read += uint32(n_saved); read >= data_length {
			return f.Name(), read, nil
		}
	}

	return "", 0, err
}
