package resolved

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

var ErrIdentityTooLarge = errors.New("resolved REST Query identity exceeds 64 KiB")

type identityWriter struct {
	buffer bytes.Buffer
	err    error
}

func (writer *identityWriter) writeByte(value byte) {
	if writer.err != nil {
		return
	}
	writer.err = writer.buffer.WriteByte(value)
	writer.checkSize()
}

func (writer *identityWriter) writeBool(value bool) {
	if value {
		writer.writeByte(1)
		return
	}
	writer.writeByte(0)
}

func (writer *identityWriter) writeUint32(value uint32) {
	if writer.err != nil {
		return
	}
	var encoded [4]byte
	binary.BigEndian.PutUint32(encoded[:], value)
	_, writer.err = writer.buffer.Write(encoded[:])
	writer.checkSize()
}

func (writer *identityWriter) writeUint64(value uint64) {
	if writer.err != nil {
		return
	}
	var encoded [8]byte
	binary.BigEndian.PutUint64(encoded[:], value)
	_, writer.err = writer.buffer.Write(encoded[:])
	writer.checkSize()
}

func (writer *identityWriter) writeBytes(value []byte) {
	if writer.err != nil {
		return
	}
	if uint64(len(value)) > uint64(^uint32(0)) {
		writer.err = ErrIdentityTooLarge
		return
	}
	writer.writeUint32(uint32(len(value)))
	if writer.err != nil {
		return
	}
	_, writer.err = writer.buffer.Write(value)
	writer.checkSize()
}

func (writer *identityWriter) writeString(value string) {
	writer.writeBytes([]byte(value))
}

func (writer *identityWriter) checkSize() {
	if writer.err == nil && writer.buffer.Len() > MaximumIdentityBytes {
		writer.err = ErrIdentityTooLarge
	}
}

func (writer *identityWriter) result() ([]byte, error) {
	if writer.err != nil {
		return nil, writer.err
	}
	if writer.buffer.Len() == 0 {
		return nil, fmt.Errorf("resolved REST Query identity is empty")
	}
	return bytes.Clone(writer.buffer.Bytes()), nil
}
