package codec

import (
	"bufio"
	"fmt"
	"github.com/pingcap/errors"
	"io"
	"reflect"
)

var _ Protocol = (*ByteProtocol)(nil)
var _ Codec = (*ByteCodec)(nil)

type ByteProtocol struct{}

func Byte() *ByteProtocol {
	return &ByteProtocol{}
}

func (bp *ByteProtocol) NewCodec(rw io.ReadWriter) (Codec, error) {
	bc := &ByteCodec{rw: bufio.NewReadWriter(bufio.NewReader(rw), bufio.NewWriter(rw))}
	bc.closer, _ = rw.(io.Closer)
	bc.source = rw
	return bc, nil
}

type ByteCodec struct {
	closer io.Closer
	rw     *bufio.ReadWriter
	source io.ReadWriter
}

func (bc *ByteCodec) HiJack() io.ReadWriter {
	return bc.source
}

func (bc *ByteCodec) Read(msg interface{}) (err error) {
	m, ok := msg.([]byte)
	if !ok {
		return fmt.Errorf("msg's type is not []byte: %v", msg)
	}
	if _, err = bc.rw.Read(m); err != nil {
		return errors.Trace(err)
	}

	reV := reflect.ValueOf(&msg).Elem()
	if reV.CanSet() {
		reV.Set(reflect.ValueOf(m))
	}
	return
}

func (bc *ByteCodec) Write(msg interface{}) (err error) {
	defer func() {
		_ = bc.rw.Flush()
		if err != nil {
			_ = bc.Close()
		}
	}()
	m, ok := msg.([]byte)
	if !ok {
		return fmt.Errorf("msg's type is not []byte: %v", msg)
	}
	if _, err := bc.rw.Write(m); err != nil {
		return errors.Trace(err)
	}
	return
}

func (bc *ByteCodec) Close() error {
	if bc.closer != nil {
		return bc.closer.Close()
	}
	return nil
}
