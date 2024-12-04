package server

import (
	"bytes"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"io"
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

func (uh *ConnectionHandler) HandleConn() error {
	var header []byte
	n, err := uh.conn.Read(header)
	if err != nil {
		return err
	}
	fmt.Println("Header size is ", n)
	err = uh.h.Deserialize(header)
	if err != nil {
		return err
	}
	fileName := filepath.Join(uh.h.Dir, uh.h.FileName)
	_, err = os.Stat(fileName)
	var fd *os.File
	if os.IsNotExist(err) {
		fd, err = os.Create(fileName)
	} else {
		fd, err = os.Open(fileName)
	}
	if err != nil {
		return err
	}
	offset := int64(0)
	for i := uint64(0); i < uh.h.Reps; i++ {
		buf := make([]byte, uh.h.ChunkSize)
		n, err := io.CopyN(bytes.NewBuffer(buf), uh.conn, int64(uh.h.ChunkSize))
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to receive bytes over tcp, err %v", err)
		}
		fmt.Println(n, "bytes were received from client")
		fd.WriteAt(buf, offset)
		offset += int64(uh.h.ChunkSize)
	}
	return nil
}
