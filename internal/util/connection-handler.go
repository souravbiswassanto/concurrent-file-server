package util

type ConnHandler interface {
	HandleConn() error
}
