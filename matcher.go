package netutil

import (
	"errors"
	"io"
	"net"
	"sync"
)

var (
	ErrByteHeaderNotFound          = errors.New("byte header not found")
	ErrByteHeaderAlreadyRegistered = errors.New("byte header already registered")
)

type Matcher struct {
	handlers      map[interface{}]RWHandler
	handlersMutex sync.Mutex
}

func CreateEmptyMatcher() *Matcher {
	return &Matcher{
		make(map[interface{}]RWHandler), sync.Mutex{},
	}
}

func (m *Matcher) Register(k interface{}, h RWHandler) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	if m.handlers[k] != nil {
		return ErrByteHeaderAlreadyRegistered
	}

	m.handlers[k] = h

	return nil
}

func (m *Matcher) RegisterBytes(bytes []byte, h RWHandler) error {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	for _, b := range bytes {
		if m.handlers[b] != nil {
			return ErrByteHeaderAlreadyRegistered
		}
	}

	for _, b := range bytes {
		m.handlers[b] = h
	}

	return nil
}

func (m *Matcher) RegisterByte(b byte, h RWHandler) error {
	return m.Register(b, h)
}

func (m *Matcher) RegisterString(s string, h RWHandler) error {
	return m.Register(s, h)
}

func (m *Matcher) Match(k interface{}) (RWHandler, error) {
	m.handlersMutex.Lock()
	defer m.handlersMutex.Unlock()

	if h := m.handlers[k]; h != nil {
		return h, nil
	}

	return nil, ErrByteHeaderNotFound
}

func (m *Matcher) Invoke(k interface{}, r io.Reader, w io.Writer, c net.Conn) error {
	handler, err := m.Match(k)
	if err != nil {
		return err
	}

	if err := handler(r, w, c); err != nil {
		return err
	}

	return nil
}
