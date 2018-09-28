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
	ErrConnectionClosed            = errors.New("connection closed")
	ErrByteHeaderNotFound          = errors.New("byte header not found")
	ErrByteHeaderAlreadyRegistered = errors.New("byte header already registered")
)

type RWHandler func(io.Reader, io.Writer, net.Conn) error

type Socket struct {
	li net.Listener

	handlers      map[byte]RWHandler
	handlersMutex sync.Mutex
}

func Bind(li net.Listener) *Socket {
	s := &Socket{
		li, make(map[byte]RWHandler), sync.Mutex{},
	}

	return s
}

func (s *Socket) Register(bytes []byte, h RWHandler) error {
	s.handlersMutex.Lock()
	defer s.handlersMutex.Unlock()

	for _, b := range bytes {
		if s.handlers[b] != nil {
			return ErrByteHeaderAlreadyRegistered
		}
	}

	for _, b := range bytes {
		s.handlers[b] = h
	}

	return nil
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

				s.handlersMutex.Lock()
				handler := s.handlers[header]
				s.handlersMutex.Unlock()

				if handler == nil {
					return ErrByteHeaderNotFound
				}

				if err := handler(r, w, conn); err != nil {
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
