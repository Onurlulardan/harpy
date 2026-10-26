package archiveartifact

import (
	"encoding/binary"
	"io"
)

func write16(target io.Writer, values ...uint16) error {
	buffer := make([]byte, len(values)*2)
	for index, value := range values {
		binary.LittleEndian.PutUint16(buffer[index*2:], value)
	}
	_, err := target.Write(buffer)
	return err
}

func write32(target io.Writer, values ...uint32) error {
	buffer := make([]byte, len(values)*4)
	for index, value := range values {
		binary.LittleEndian.PutUint32(buffer[index*4:], value)
	}
	_, err := target.Write(buffer)
	return err
}

func write64(target io.Writer, values ...uint64) error {
	buffer := make([]byte, len(values)*8)
	for index, value := range values {
		binary.LittleEndian.PutUint64(buffer[index*8:], value)
	}
	_, err := target.Write(buffer)
	return err
}
