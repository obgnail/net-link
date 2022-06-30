package codec

import (
	"bufio"
	"encoding/json"
	"github.com/juju/errors"
	"io"
)

var _ Protocol = (*JsonProtocol)(nil)
var _ Codec = (*JsonCodec)(nil)

type JsonProtocol struct{}

func Json() *JsonProtocol {
	return &JsonProtocol{}
}

func (jp *JsonProtocol) NewCodec(rw io.ReadWriter) (Codec, error) {
	buf := bufio.NewWriter(rw)
	jc := &JsonCodec{
		buf:     buf,
		encoder: json.NewEncoder(buf),
		decoder: json.NewDecoder(rw),
		source:  rw,
	}
	jc.closer, _ = rw.(io.Closer) // rw不一定实现了closer接口
	return jc, nil
}

type JsonCodec struct {
	closer  io.Closer
	buf     *bufio.Writer
	encoder *json.Encoder
	decoder *json.Decoder
	source  io.ReadWriter
}

func (jc *JsonCodec) HiJack() io.ReadWriter {
	return jc.source
}

func (jc *JsonCodec) Read(msg interface{}) (err error) {
	if err := jc.decoder.Decode(msg); err != nil {
		return errors.Trace(err)
	}
	return
}

func (jc *JsonCodec) Write(msg interface{}) (err error) {
	defer func() {
		_ = jc.buf.Flush()
		if err != nil {
			_ = jc.Close()
		}
	}()

	if err = jc.encoder.Encode(&msg); err != nil {
		return errors.Trace(err)
	}
	return
}

func (jc *JsonCodec) Close() error {
	if jc.closer != nil {
		return jc.closer.Close()
	}
	return nil
}
