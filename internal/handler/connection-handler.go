package handler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

type ConnectionHandler struct {
	conn *net.TCPConn
	h    *util.Header
}

func NewConnectionHandler(conn *net.TCPConn, h *util.Header) *ConnectionHandler {
	return &ConnectionHandler{
		conn: conn,
		h:    h,
	}

}

func (h *ConnectionHandler) HandleConn() error {
	err := h.Handshake()
	if err != nil {
		return err
	}
	err = h.HandleHeader()
	if err != nil {
		return err
	}
	fd, err := h.HandleFile()
	if err != nil {
		return err
	}
	defer func() {
		err = fd.Close()
		if err != nil {
			log.Fatalln(err)
		}
	}()
	return h.HandleReceive(fd)
}

func (h *ConnectionHandler) Handshake() error {
	temp := make([]byte, 4)
	n, err := h.conn.Read(temp)
	if err != nil {
		return err
	}
	log.Println(n, "bytes initial data received")
	if n != 4 {
		return fmt.Errorf("not a valid header")
	}
	_, err = h.conn.Write(temp)
	return err
}

func (h *ConnectionHandler) HandleHeader() error {
	headerLen := make([]byte, 4)
	_, err := h.conn.Read(headerLen)
	if err != nil {
		return err
	}
	hLen := binary.BigEndian.Uint32(headerLen[:])
	var header bytes.Buffer
	_, err = io.CopyN(&header, h.conn, int64(hLen))
	if err != nil {
		return err
	}
	return h.h.Deserialize(header.Bytes())
}

func (h *ConnectionHandler) HandleFile() (*os.File, error) {
	fileName := filepath.Join(h.h.Dir, h.h.FileName)
	fileName = filepath.Join("storage", fileName)
	fmt.Println(fileName, h.h.ChunkSize)
	var fd *os.File
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		dir := filepath.Dir(fileName)
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			return nil, err
		}
		fd, err = os.Create(fileName)
		if err != nil {
			return nil, err
		}
	} else if err == nil {
		fd, err = os.OpenFile(fileName, os.O_RDWR, 0775)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, err
	}
	log.Println(fd.Name())
	return fd, nil
}

func (h *ConnectionHandler) HandleReceive(fd *os.File) error {
	offset := int64(0)
	var sz uint64
	for i := uint64(0); i <= h.h.Reps; i++ {
		sz = uint64(h.h.ChunkSize)
		if i == h.h.Reps {
			sz = h.h.FileSize % uint64(h.h.ChunkSize)
		}
		var temp bytes.Buffer
		n, err := io.CopyN(&temp, h.conn, int64(sz))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to receive bytes over tcp, err %v", err)
			return err
		}
		fmt.Println(n, "bytes were received from client")
		_, err = fd.WriteAt(temp.Bytes(), offset)
		if err != nil {
			return err
		}
		offset += int64(sz)
		// time.Sleep(time.Millisecond * 200)
	}
	return nil
}
