package server

import (
	"context"
	"encoding/binary"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type HandleFunc func(conn *net.TCPConn) error

type FileServer struct {
	listener *net.TCPListener
	port     string
	ip       string
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

func NewFileServer(ctx context.Context, port, ip string) FileServer {
	ctx, cancel := context.WithCancel(ctx)
	return FileServer{
		port:   port,
		ip:     ip,
		ctx:    ctx,
		cancel: cancel,
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
	return err
}

func (fs *FileServer) Start() error {
	return fs.setupTCPListener()
}

func (fs *FileServer) Shutdown() {
	if fs.listener != nil {
		fs.listener.Close()
	}
	fs.cancel()
	fs.wg.Wait()
}

func (fs *FileServer) Run(handleFn HandleFunc) {
	errStream := make(chan error, 1)
	sigStream := make(chan os.Signal, 1)
	signal.Notify(sigStream, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			select {
			case <-fs.ctx.Done():
				return
			case <-sigStream:
				return
			case err := <-errStream:
				log.Println(err)
			}
		}
	}()

	for {
		select {
		case <-fs.ctx.Done():
			return
		case <-sigStream:
			return
		default:
			conn, err := fs.listener.AcceptTCP()
			if err != nil {
				errStream <- err
				continue
			}

			fs.wg.Add(1)
			go func(tc *net.TCPConn) {
				select {
				case <-fs.ctx.Done():
					return
				default:
					err := handleFn(tc)
					if err != nil {
						errStream <- err
					}
				}
				defer fs.wg.Done()
			}(conn)
		}
	}
}

func handleConn(conn *net.TCPConn) error {
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
	srv := NewFileServer(context.Background(), "127.0.0.1", "8000")
	err := srv.Start()
	if err != nil {
		return srv, err
	}
	return srv, nil
}

func SetupAndRunServer(fn HandleFunc) error {
	srv, err := SetupServer()
	if err != nil {
		return err
	}
	srv.Run(fn)
	return nil
}
