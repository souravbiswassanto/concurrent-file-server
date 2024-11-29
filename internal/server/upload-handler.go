package server

import "github.com/souravbiswassanto/concurrent-file-server/internal/util"

type ConnectionHandler struct {
	fs *FileServer
	h  *util.Header
}

func (uh *ConnectionHandler) HandleConn() {

}
