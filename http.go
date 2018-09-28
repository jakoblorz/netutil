package netutil

import (
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

func ServeMuxRWHandler(setupFunc func(*http.ServeMux), addr net.Addr) RWHandler {

	m := http.NewServeMux()
	setupFunc(m)

	ln := &inplaceTCPListener{
		c:    make(chan net.Conn),
		addr: addr,
	}

	go http.Serve(ln, m)

	return func(r io.Reader, _ io.Writer, conn net.Conn) error {
		ln.c <- &inplaceTCPConnection{
			conn: conn,
			r:    r,
		}

		return nil
	}
}

type inplaceTCPListener struct {
	c    chan net.Conn
	once sync.Once
	addr net.Addr
}

func (i *inplaceTCPListener) Accept() (c net.Conn, err error) {
	conn, ok := <-i.c
	if !ok {
		return nil, ErrConnectionClosed
	}
	return conn, nil
}

func (i *inplaceTCPListener) Close() error {
	i.once.Do(func() { close(i.c) })
	return nil
}

func (i *inplaceTCPListener) Addr() net.Addr {
	return i.addr
}

type inplaceTCPConnection struct {
	conn net.Conn
	r    io.Reader
}

func (c *inplaceTCPConnection) Read(b []byte) (n int, err error)   { return c.r.Read(b) }
func (c *inplaceTCPConnection) Write(b []byte) (n int, err error)  { return c.conn.Write(b) }
func (c *inplaceTCPConnection) Close() error                       { return c.conn.Close() }
func (c *inplaceTCPConnection) LocalAddr() net.Addr                { return c.conn.LocalAddr() }
func (c *inplaceTCPConnection) RemoteAddr() net.Addr               { return c.conn.RemoteAddr() }
func (c *inplaceTCPConnection) SetDeadline(t time.Time) error      { return c.conn.SetDeadline(t) }
func (c *inplaceTCPConnection) SetReadDeadline(t time.Time) error  { return c.conn.SetReadDeadline(t) }
func (c *inplaceTCPConnection) SetWriteDeadline(t time.Time) error { return c.conn.SetWriteDeadline(t) }
