package netutil

import (
	"bufio"
	"errors"
	"io"
	"net"

	"github.com/ugorji/go/codec"
)

var (
	ErrWrongHandlerReached = errors.New("wrong handler reached")
)

type BytesRWHandlerFunc func(interface{}) (interface{}, error)

func BytesRWHandler(handle BytesRWHandlerFunc, i interface{}, c codec.Handle) RWHandler {
	return func(ir io.Reader, w io.Writer, _ net.Conn) error {

		r := bufio.NewReader(ir)
		_, err := r.ReadByte()
		if err != nil {
			return err
		}

		dec := codec.NewDecoder(r, c)
		enc := codec.NewEncoder(w, c)
		req := func(i interface{}) interface{} {
			return i
		}(i)

		if err := dec.Decode(&req); err != nil {
			return err
		}

		res, e := func(res interface{}, err error) (interface{}, string) {
			str := ""
			if err != nil {
				str = err.Error()
			}

			return res, str
		}(handle(req))

		if err := enc.Encode(e); err != nil {
			return err
		}

		if err := enc.Encode(res); err != nil {
			return err
		}

		return nil
	}
}
