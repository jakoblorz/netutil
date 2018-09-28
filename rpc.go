package netutil

import (
	"io"
	"net"
	"net/rpc"
)

func GoNetRPCRWHandler(setupFunc func(*rpc.Server)) RWHandler {

	r := rpc.NewServer()
	setupFunc(r)

	return func(_ io.Reader, _ io.Writer, conn net.Conn) error {
		r.ServeConn(conn)
		return nil
	}
}
