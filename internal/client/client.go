package client

import (
	"context"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/cmd/client"
	"github.com/souravbiswassanto/concurrent-file-server/internal/tcp"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"log"
	"net"
	"os"
	"sync"
)

type FileClient struct {
	ctx          context.Context
	wg           sync.WaitGroup
	cAddr, sAddr *net.TCPAddr
	ch           util.ConnHandler
}

func NewFileClient(ctx context.Context, cIP, cPort, sIP, sPort string) (*FileClient, error) {
	fc := FileClient{}
	fc.ctx = ctx
	cAddr, err := fc.clientAddr(cIP, cPort)
	if err != nil {
		return nil, err
	}
	sAddr, err := fc.resolveServerTcpAddr(sIP, sPort)
	if err != nil {
		return nil, err
	}
	fc.cAddr = cAddr
	fc.sAddr = sAddr
	return &fc, nil
}

func (fs *FileClient) resolveServerTcpAddr(sIP, sPort string) (*net.TCPAddr, error) {
	if sIP == "" || sPort == "" {
		sIP = os.Getenv("SERVER_IP")
		sPort = os.Getenv("SERVER_PORT")
	}
	if sIP == "" || sPort == "" {
		return nil, fmt.Errorf("either provide -P port -I ip or set SERVER_IP SERVER_PORT")
	}
	return net.ResolveTCPAddr("tcp", sIP+":"+sPort)
}

func (fs *FileClient) resolveTcpAddr(cIP, cPort string) (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", cIP+":"+cPort)
}

func (fc *FileClient) clientAddr(cIP, cPort string) (*net.TCPAddr, error) {
	if cIP == "" || cPort == "" {
		log.Println("client port or IP not given, using defaults")
		return nil, nil
	}
	return fc.resolveTcpAddr(cIP, cPort)
}

func (fc *FileClient) DialTCPWithContext() (*net.TCPConn, error) {
	connStream := make(chan *net.TCPConn)
	errStream := make(chan error)
	// this part of the code is inspired from
	// github.com/kubedb/pg-coordinator/pkg/listener.go

	go func() {
		dialer := &net.Dialer{}
		conn, err := dialer.DialContext(fc.ctx, "tcp", fc.sAddr.String())
		if err != nil {
			errStream <- err
			return
		}
		tcpConn, ok := conn.(*net.TCPConn)
		if !ok {
			conn.Close()
			errStream <- fmt.Errorf("unexpected connection type")
		}
		connStream <- tcpConn
	}()

	select {
	case <-fc.ctx.Done():
		return nil, fmt.Errorf("connection dropped")
	case t := <-connStream:
		return t, nil
	case err := <-errStream:
		return nil, err
	}
}

//func (fc *FileClient) handleConn(conn *net.TCPConn) {
//	//fd, err := os.Open("internal/client/client.go")
//	//if err != nil {
//	//	log.Println("err occoured: ", err)
//	//	return
//	//}
//	//data, err := io.ReadAll(fd)
//	//if err != nil {
//	//	log.Println("err occoured: ", err)
//	//	return
//	//}
//	//n, err := conn.Write(data)
//	//if err != nil {
//	//	log.Println("err: ", err)
//	//	return
//	//}
//	//log.Println("sent ", n, "bytes over network")
//
//}

func GetHandler(ctx context.Context, uc client.UploadConfig) (util.ConnHandler, error) {

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

func HandleUpload(uc client.UploadConfig) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	h, err := GetHandler(ctx, uc)
	if err != nil {
		return err
	}

	return h.HandleConn()
}
