package client

import (
	"context"
	"fmt"
	"github.com/souravbiswassanto/concurrent-file-server/internal/util"
	"net"
	"os"
	"sync"
)

type FileClient struct {
	ctx                            context.Context
	wg                             sync.WaitGroup
	ip, serverIP, port, serverPort string
	conn                           *net.TCPConn
}

func NewFileClient(ip, port string) *FileClient {
	fc := FileClient{}
	fc.Setup(ip, port)
	return &fc
}

func (fc *FileClient) Setup(ip, port string) {
	fc.ip = ip
	fc.port = port
	fc.SetupServer()
}

func (fc *FileClient) SetupServer() {
	serverIP := os.Getenv("SERVER_IP")
	serverPort := os.Getenv("SERVER_PORT")
	//if serverIP == "" || serverPort == "" {
	//	log.Fatalln("Provide SERVER_IP and SERVER_PORT env variables")
	//}
	fc.serverIP = serverIP
	fc.serverPort = serverPort
}

func (fs *FileClient) resolveServerTcpAddr() (*net.TCPAddr, error) {
	return net.ResolveTCPAddr("tcp", fs.serverIP+":"+fs.serverPort)
}

func (fs *FileClient) resolveTcpAddr() (*net.TCPAddr, error) {
	if fs.ip == "" {
		fs.ip = "127.0.0.1"
	}
	if fs.port == "" {
		fs.port = "8080"
	}
	return net.ResolveTCPAddr("tcp", fs.ip+":"+fs.port)
}

func (fc *FileClient) clientAddr() (*net.TCPAddr, error) {
	cip := fc.ip
	cport := fc.port
	if cip == "" || cport == "" {
		return nil, nil
	}
	clientAddr, err := fc.resolveTcpAddr()
	if err != nil {
		return nil, err
	}
	return clientAddr, nil
}

func (fc *FileClient) Start() error {
	cAddr, err := fc.clientAddr()
	if err != nil {
		return err
	}
	serverAddr, err := fc.resolveServerTcpAddr()
	if err != nil {
		return err
	}
	conn, err := net.DialTCP("tcp", cAddr, serverAddr)
	if err != nil {
		return err
	}
	fc.conn = conn
	return nil
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

func (fc *FileClient) GetConnection() (*net.TCPConn, error) {
	if fc.conn == nil {
		return nil, fmt.Errorf("the connections is closed, start a new connection by running Start()")
	}
	return fc.conn, nil
}

func (fc *FileClient) HandleUpload(uh util.HandleFunc) {

}
