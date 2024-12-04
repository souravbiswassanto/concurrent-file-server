package tcp

import (
	"bytes"
	"context"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/client"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"io"
	"log"
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

	_, err = io.Copy(conn, bytes.NewReader(uh.h.Serialize()))
	if err != nil {
		return err
	}
	// receive ack
	temp := make([]byte, 4)
	_, err = io.Copy(bytes.NewBuffer(temp), conn)
	if err != nil {
		return err
	}
	if !validateAck(temp) {
		return fmt.Errorf("invalid response received from server")
	}

	offset := uint64(0)
	for i := uint64(0); i <= uh.h.Reps; i++ {
		buf := make([]byte, uh.h.ChunkSize)
		n, err := fd.ReadAt(buf, int64(offset))
		if err != nil && err != io.EOF {
			return err
		}
		_, err = io.Copy(conn, bytes.NewReader(buf))
		if err != nil {
			return err
		}
		log.Println(n, "bytes sent to server over network")
	}
	return nil
}

func validateAck(res []byte) bool {
	if len(res) != 4 {
		return false
	}
	for i := 0; i < 4; i++ {
		if res[i] != byte(1) {
			return false
		}
	}
	return true
}
