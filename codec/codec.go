package codec

import (
	"io"
)

type Codec interface {
	io.Closer
	Read(receiver interface{}) error
	Write(send interface{}) error
	HiJack() io.ReadWriter
}

type Protocol interface {
	// 并不要求rw实现close接口
	NewCodec(rw io.ReadWriter) (Codec, error)
}
