package tcp

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/client"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

type UploadHandler struct {
	fc *client.FileClient
	h  *util.Header
}

func NewUploadHandler(ctx context.Context, uc util.UploadConfig) (*UploadHandler, error) {
	fc, err := client.NewFileClient(ctx, uc.CIP, uc.CPort, uc.SIP, uc.SPort)
	if err != nil {
		return nil, err
	}
	h, err := util.NewHeader(uc.File, uint32(uc.ChunkSize))
	if err != nil {
		return nil, err
	}
	return &UploadHandler{
		fc: fc,
		h:  h,
	}, nil
}

func (uh *UploadHandler) HandleConn() error {

	fname := filepath.Join(uh.h.Dir, uh.h.FileName)
	fd, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("failed to open %s, err: %v", fname, err)
	}
	conn, err := uh.fc.DialTCPWithContext()
	if err != nil {
		return err
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	err = uh.Handshake(conn)
	if err != nil {
		return err
	}

	err = uh.SetupHeader(conn)
	if err != nil {
		return err
	}

	return uh.HandleSend(conn, fd)
}

func (uh *UploadHandler) Handshake(conn *net.TCPConn) error {
	temp := []byte{1, 1, 1, 1}
	_, err := conn.Write(temp)
	if err != nil {
		return err
	}
	temp = make([]byte, 4)
	n, err := conn.Read(temp)
	if err != nil {
		return err
	}
	if n != 4 {
		return fmt.Errorf("wrong header received")
	}
	return nil
}

func (uh *UploadHandler) SetupHeader(conn *net.TCPConn) error {
	headerBuf := make([]byte, 4)
	var buf []byte
	buf = uh.h.Serialize()
	binary.BigEndian.PutUint32(headerBuf, uint32(len(buf)))
	_, err := conn.Write(headerBuf)
	if err != nil {
		return err
	}
	fmt.Println(len(buf), uh.h.FileSize)
	_, err = io.CopyN(conn, bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return err
	}
	return nil
}

func (uh *UploadHandler) HandleSend(conn *net.TCPConn, fd *os.File) error {
	offset := uint64(0)
	sz := uint64(0)
	for i := uint64(0); i <= uh.h.Reps; i++ {
		sz = uint64(uh.h.ChunkSize)
		if i == uh.h.Reps {
			sz = uh.h.FileSize % uint64(uh.h.ChunkSize)
		}
		buf := make([]byte, sz)
		n, err := fd.ReadAt(buf, int64(offset))
		if err != nil && err != io.EOF {
			return err
		}
		_, err = io.CopyN(conn, bytes.NewReader(buf), int64(sz))
		if err != nil {
			return err
		}
		log.Println(n, "bytes sent to server over network")
		offset += sz
	}
	return nil
}
