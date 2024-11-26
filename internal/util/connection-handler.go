package util

import (
	"context"
	"net"
)

type ConnHandler interface {
	HandleConn(ctx context.Context, conn *net.TCPConn) error
}

type HandleFunc struct {
}

func (hf HandleFunc) HandleConn(ctx context.Context, conn *net.TCPConn) error {
	var buf []byte
	_, err := conn.Read(buf)
	if err != nil {
		return err
	}

	return nil
}
