package handler

import (
	"context"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"github.com/souravbiswassanto/concurrent-file-server/protocol/client/tcp"
)

//type UploadHandler struct {
//	fc *client.FileClient
//	h  *util.Header
//}
//
//func (uh *UploadHandler) HandleConn() error {
//	fname := filepath.Join(uh.h.Dir, uh.h.FileName)
//	fd, err := os.Open(fname)
//	if err != nil {
//		return fmt.Errorf("failed to open %s, err: %v", fname, err)
//	}
//	conn, err := uh.fc.DialTCPWithContext()
//	if err != nil {
//		return err
//	}
//	defer func() {
//		err = conn.Close()
//		if err != nil {
//			log.Println(err)
//		}
//	}()
//
//	_, err = io.Copy(conn, bytes.NewReader(uh.h.Serialize()))
//	if err != nil {
//		return err
//	}
//	//// receive ack
//	//temp := make([]byte, 4)
//	//_, err = io.Copy(bytes.NewBuffer(temp), conn)
//	//if err != nil {
//	//	return err
//	//}
//	//if !validateAck(temp) {
//	//	return fmt.Errorf("invalid response received from server")
//	//}
//
//	offset := uint64(0)
//	sz := uint64(0)
//	for i := uint64(0); i <= uh.h.Reps; i++ {
//		sz = uint64(uh.h.ChunkSize)
//		if i == uh.h.Reps {
//			sz = uh.h.FileSize % uint64(uh.h.ChunkSize)
//		}
//		buf := make([]byte, sz)
//		n, err := fd.ReadAt(buf, int64(offset))
//		if err != nil && err != io.EOF {
//			return err
//		}
//		log.Println(n, err)
//		_, err = io.CopyN(conn, bytes.NewReader(buf), int64(sz))
//		if err != nil {
//			return err
//		}
//		log.Println(n, "bytes sent to server over network")
//	}
//	return nil
//}
//
//func validateAck(res []byte) bool {
//	if len(res) != 4 {
//		return false
//	}
//	for i := 0; i < 4; i++ {
//		if res[i] != byte(1) {
//			return false
//		}
//	}
//	return true
//}

func GetHandler(ctx context.Context, uc util.UploadConfig) (util.ConnHandler, error) {

	switch uc.Protocol {
	case "tcp":
		th, err := tcp.NewUploadHandler(ctx, uc)
		if err != nil {
			return nil, err
		}
		return th, nil
	}
	return nil, fmt.Errorf("no connection type matched")
}

func HandleUpload(uc util.UploadConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h, err := GetHandler(ctx, uc)
	if err != nil {
		return err
	}

	return h.HandleConn()
}
