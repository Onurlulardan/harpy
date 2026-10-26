package notificationhttp

import (
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"net"
)

const (
	socketOpcodeText  byte = 1
	socketOpcodeClose byte = 8
	socketOpcodePing  byte = 9
	socketOpcodePong  byte = 10
	maximumFrameBytes      = 4096
)

var errSocketFrame = errors.New("notification websocket frame is invalid")

func readClientFrame(reader *bufio.Reader) (byte, []byte, error) {
	var header [2]byte
	if _, err := io.ReadFull(reader, header[:]); err != nil {
		return 0, nil, err
	}
	fin, masked, opcode := header[0]&0x80 != 0, header[1]&0x80 != 0, header[0]&0x0f
	if !fin || !masked || header[0]&0x70 != 0 || !validClientOpcode(opcode) {
		return 0, nil, errSocketFrame
	}
	length, err := readFrameLength(reader, header[1]&0x7f)
	if err != nil || length > maximumFrameBytes || opcode >= 8 && length > 125 {
		return 0, nil, errSocketFrame
	}
	var mask [4]byte
	if _, err = io.ReadFull(reader, mask[:]); err != nil {
		return 0, nil, err
	}
	payload := make([]byte, length)
	if _, err = io.ReadFull(reader, payload); err != nil {
		return 0, nil, err
	}
	for index := range payload {
		payload[index] ^= mask[index%len(mask)]
	}
	return opcode, payload, nil
}

func readFrameLength(reader *bufio.Reader, marker byte) (uint64, error) {
	if marker < 126 {
		return uint64(marker), nil
	}
	if marker == 126 {
		var value uint16
		err := binary.Read(reader, binary.BigEndian, &value)
		return uint64(value), err
	}
	var value uint64
	err := binary.Read(reader, binary.BigEndian, &value)
	return value, err
}

func writeServerFrame(connection net.Conn, opcode byte, payload []byte) error {
	if len(payload) > maximumFrameBytes || !validServerOpcode(opcode) {
		return errSocketFrame
	}
	header := []byte{0x80 | opcode}
	switch {
	case len(payload) < 126:
		header = append(header, byte(len(payload)))
	default:
		header = append(header, 126, byte(len(payload)>>8), byte(len(payload)))
	}
	if err := writeFrameBytes(connection, header); err != nil {
		return err
	}
	return writeFrameBytes(connection, payload)
}

func writeFrameBytes(target io.Writer, value []byte) error {
	for len(value) != 0 {
		written, err := target.Write(value)
		if err != nil {
			return err
		}
		if written < 1 || written > len(value) {
			return io.ErrShortWrite
		}
		value = value[written:]
	}
	return nil
}

func validClientOpcode(value byte) bool {
	return value == socketOpcodeText || value == socketOpcodeClose ||
		value == socketOpcodePing || value == socketOpcodePong
}

func validServerOpcode(value byte) bool {
	return value == socketOpcodeText || value == socketOpcodeClose || value == socketOpcodePong
}
