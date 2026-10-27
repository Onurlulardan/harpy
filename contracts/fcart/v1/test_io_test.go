package fcartv1

import (
	"bytes"
	"io"
)

type shortBuffer struct {
	bytes.Buffer
	short bool
}

func (buffer *shortBuffer) Write(value []byte) (int, error) {
	if buffer.short && len(value) > 0 {
		return buffer.Buffer.Write(value[:len(value)-1])
	}
	return buffer.Buffer.Write(value)
}

func newByteReader(value []byte) io.Reader {
	return bytes.NewReader(value)
}
