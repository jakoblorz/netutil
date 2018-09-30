package netutil

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

var (
	ErrConnectionClosed = errors.New("connection closed")
)

type RWHandler func(io.Reader, io.Writer, net.Conn) error

type Socket struct {
	li net.Listener

	matcher *Matcher
}

func Bind(li net.Listener) *Socket {
	s := &Socket{
		li, CreateEmptyMatcher(),
	}

	return s
}

func (s *Socket) Matcher() *Matcher {
	return s.matcher
}

func (s *Socket) Register(bytes []byte, h RWHandler) error {
	return s.matcher.RegisterBytes(bytes, h)
}

func (s *Socket) Accept(wg *sync.WaitGroup, lg *log.Logger, t time.Duration) error {
	for {
		conn, err := s.li.Accept()
		if err != nil {
			return err
		}

		wg.Add(1)
		go func(conn net.Conn) {
			defer wg.Done()

			err := func() error {

				r := bufio.NewReader(conn)
				w := bufio.NewWriter(conn)

				if err := conn.SetReadDeadline(time.Now().Add(t)); err != nil {
					return err
				}

				header, err := r.ReadByte()
				if err != nil {
					return err
				} else if err := r.UnreadByte(); err != nil {
					return err
				}

				if err := conn.SetReadDeadline(time.Time{}); err != nil {
					return err
				}

				if err := s.matcher.Invoke(header, r, w, conn); err != nil {
					return err
				}

				return nil
			}()

			if err != nil {
				conn.Close()
				lg.Printf("error during connection handling: %v\n", err)
			}
		}(conn)
	}
}
