package util

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
)

type ConnHandler interface {
	HandleConn(ctx context.Context, conn *net.TCPConn) error
}

type HandleFunc struct {
}

func (hf HandleFunc) HandleConn(ctx context.Context, conn *net.TCPConn) error {
	for i := 0; i < 1717/128+1; i++ {
		buf := make([]byte, 128)
		n, err := io.CopyN(bytes.NewBuffer(buf), conn, 128)
		if err != nil {
			return err
		}
		log.Println(buf)
		log.Println(n, "bytes are rate")
	}
	return nil
}
