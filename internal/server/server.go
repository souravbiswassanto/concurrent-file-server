package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/souravbiswassanto/concurrent-file-server/internal/handler"
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
	isServerRunning *bool
}

func NewFileServer(ctx context.Context, ip, port string) FileServer {
	ctx, cancel := context.WithCancel(ctx)
	return FileServer{
		port:            port,
		ip:              ip,
		ctx:             ctx,
		cancel:          cancel,
		isServerRunning: nil,
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
	if err != nil {
		return err
	}
	fs.listener, err = net.ListenTCP("tcp", addr)
	if err == nil {
		fs.isServerRunning = aws.Bool(true)
		log.Println("listening at:", addr.String())
	}
	return err
}

func (fs *FileServer) Start() error {
	return fs.setupTCPListener()
}

func (fs *FileServer) IsServerRunning() bool {
	if fs.isServerRunning != nil && *fs.isServerRunning {
		return true
	}
	return false
}

func (fs *FileServer) Shutdown() {
	fs.mu.Lock()
	if fs.isServerRunning != nil && !*fs.isServerRunning {
		return
	}
	if fs.listener != nil {
		fs.listener.Close()
	}
	fs.cancel()
	log.Println("Got shutdown request. Waiting for active processes to shutdown.")
	fs.wg.Wait()
	log.Println("All the process cleaned properly. Shutting Down")
	fs.isServerRunning = aws.Bool(false)
	fs.mu.Unlock()
}

func (fs *FileServer) Run() {
	errStream := make(chan error, 1)
	sigStream := make(chan os.Signal, 1)
	signal.Notify(sigStream, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGABRT)
	go func() {
		for {
			select {
			case <-fs.ctx.Done():
				fs.Shutdown()
				return
			case <-sigStream:
				fs.Shutdown()
				return
			case err := <-errStream:
				log.Println(err)
			}
		}
	}()

	for {
		conn, err := fs.listener.AcceptTCP()
		if err != nil {
			var opErr *net.OpError
			if errors.As(err, &opErr) && errors.Is(opErr.Err, net.ErrClosed) {
				log.Println("Listener was closed")
				fs.Shutdown()
				break
			}
			if conn != nil {
				errStream <- fmt.Errorf("can't accept the connection from %v, err: %v", conn.RemoteAddr(), err)
			} else {
				errStream <- err
			}
			continue
		}
		fs.wg.Add(1)
		go fs.handleConn(conn, errStream)
	}
}

func (fs *FileServer) handleConn(conn *net.TCPConn, errStream chan error) {
	workDone := make(chan struct{})
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
	h := handler.NewConnectionHandler(conn, &util.Header{})
	err := h.HandleConn()
	if err != nil {
		errStream <- fmt.Errorf("err handling connection %v", err)
	}
	workDone <- struct{}{}
}

func SetupServer() (FileServer, error) {
	srv := NewFileServer(context.Background(), readIP(), readPort())
	err := srv.Start()
	if err != nil {
		return srv, err
	}
	return srv, nil
}

func SetupAndRunServer() error {
	srv, err := SetupServer()
	if err != nil {
		return err
	}
	defer srv.Shutdown()
	srv.Run()
	return nil
}

func readIP() string {
	return os.Getenv("SERVER_IP")
}
func readPort() string {
	return os.Getenv("SERVER_PORT")
}
