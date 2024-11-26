package client

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type FileClient struct {
	ctx                            context.Context
	wg                             sync.WaitGroup
	ip, serverIP, port, serverPort string
}

func (fc *FileClient) DefaultSetup() {
	fc.Setup("127.0.0.1", "8081")
}

func (fc *FileClient) Setup(ip, port string) {
	fc.ip = ip
	fc.port = port
	fc.SetupServer()
}

func (fc *FileClient) SetupServer() {
	serverIP := os.Getenv("SERVER_IP")
	serverPort := os.Getenv("SERVER_PORT")
	if serverIP == "" || serverPort == "" {
		log.Fatalln("Provide SERVER_IP and SERVER_PORT env variables")
	}
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

func (fc *FileClient) Start() {
	clientAddr, err := fc.resolveTcpAddr()
	if err != nil {
		log.Fatalln(err)
	}
	serverAddr, err := fc.resolveServerTcpAddr()
	if err != nil {
		log.Fatalln(err)
	}
	conn, err := net.DialTCP("tcp", clientAddr, serverAddr)
	if err != nil {
		log.Fatalln(err)
	}
	handleConn(conn)
}

func handleConn(conn *net.TCPConn) {
	fd, err := os.Open("internal/client/client.go")
	if err != nil {
		log.Println("err occoured: ", err)
		return
	}
	data, err := io.ReadAll(fd)
	if err != nil {
		log.Println("err occoured: ", err)
		return
	}
	n, err := conn.Write(data)
	if err != nil {
		log.Println("err: ", err)
		return
	}
	log.Println("sent ", n, "bytes over network")
}
