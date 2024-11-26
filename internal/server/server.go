package server

import (
	"context"
	"encoding/binary"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type FileServer struct {
	listener        *net.TCPListener
	port            string
	ip              string
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	isServerRunning bool
}

func NewFileServer(ctx context.Context, ip, port string) FileServer {
	ctx, cancel := context.WithCancel(ctx)
	return FileServer{
		port:            port,
		ip:              ip,
		ctx:             ctx,
		cancel:          cancel,
		isServerRunning: true,
	}
}

func (fs *FileServer) resolveTcpAddr() (*net.TCPAddr, error) {
	if fs.ip == "" {
		fs.ip = "127.0.0.1"
	}
	if fs.port == "" {
		fs.port = "8080"
	}
	return net.ResolveTCPAddr("tcp", fs.ip+":"+fs.port)
}

func (fs *FileServer) setupTCPListener() error {
	addr, err := fs.resolveTcpAddr()
	log.Println(err)
	if err != nil {
		return err
	}
	log.Println(addr.String())
	fs.listener, err = net.ListenTCP("tcp", addr)
	return err
}

func (fs *FileServer) Start() error {
	return fs.setupTCPListener()
}

func (fs *FileServer) Shutdown() {
	fs.mu.Lock()
	if !fs.isServerRunning {
		return
	}
	if fs.listener != nil {
		fs.listener.Close()
	}
	fs.cancel()
	log.Println("Got shutdown request. Waiting for active processes to shutdown.")
	fs.wg.Wait()
	log.Println("All the process cleaned properly. Shutting Down")
	fs.isServerRunning = false
	fs.mu.Unlock()
}

func (fs *FileServer) Run(ch util.ConnHandler) {
	errStream := make(chan error, 1)
	sigStream := make(chan os.Signal, 1)
	signal.Notify(sigStream, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		select {
		case <-fs.ctx.Done():
			fs.Shutdown()
			return
		case <-sigStream:
			fs.Shutdown()
			return
		}
	}()
	go func() {
		for {
			select {
			case err := <-errStream:
				log.Println(err)
			}
		}
	}()

	for {

		log.Println(1)
		conn, err := fs.listener.AcceptTCP()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Err == net.ErrClosed {
				log.Println("Listener was closed")
				break
			}
			errStream <- err
			continue
		}
		log.Println(2)
		fs.wg.Add(1)
		go fs.handleConn(conn, ch, errStream)
	}
}

func (fs *FileServer) handleConn(conn *net.TCPConn, ch util.ConnHandler, errStream chan error) {
	workDone := make(chan bool)
	go func() {
		select {
		case <-fs.ctx.Done():
			conn.Close()
			fs.wg.Done()
			return
		case <-workDone:
			conn.Close()
			fs.wg.Done()
			return
		}
	}()
	err := ch.HandleConn(fs.ctx, conn)
	if err != nil {
		errStream <- fmt.Errorf("err handling connection %v", err)
	}
	<-workDone
}

func sampleHandleConn(conn *net.TCPConn) error {
	var sz uint32
	defer conn.Close()
	temp := make([]byte, 4)
	n, err := conn.Read(temp)
	if err != nil {
		return err
	}
	sz = binary.BigEndian.Uint32(temp[:])
	log.Println(n, sz)
	_, err = conn.Write([]byte("string received"))
	if err != nil {
		return err
	}
	return nil
}

func SetupServer() (FileServer, error) {
	srv := NewFileServer(context.Background(), "127.0.0.1", "8080")
	err := srv.Start()
	if err != nil {
		return srv, err
	}
	return srv, nil
}

func SetupAndRunServer(ch util.ConnHandler) error {
	srv, err := SetupServer()
	if err != nil {
		return err
	}
	defer srv.Shutdown()
	srv.Run(ch)
	return nil
}
