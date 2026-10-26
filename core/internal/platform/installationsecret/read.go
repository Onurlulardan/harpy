package installationsecret

import (
	"io"
	"os"
)

func Read(path string, consumer func([]byte) error) error {
	if consumer == nil {
		return ErrInvalid
	}
	target, err := validatedTarget(path, true)
	if err != nil {
		return err
	}
	file, err := openOwnerOnly(target)
	if err != nil {
		return ErrFenced
	}
	defer file.Close()
	initial, err := file.Stat()
	if err != nil || !initial.Mode().IsRegular() || initial.Size() != wirePasswordBytes {
		return ErrFenced
	}
	wire := make([]byte, wirePasswordBytes)
	_, readErr := io.ReadFull(file, wire)
	var extra [1]byte
	count, trailingErr := file.Read(extra[:])
	final, statErr := file.Stat()
	if readErr != nil || count != 0 || trailingErr != io.EOF || statErr != nil ||
		!os.SameFile(initial, final) || initial.Size() != final.Size() ||
		!initial.ModTime().Equal(final.ModTime()) || !validPasswordWire(wire) {
		clear(wire)
		return ErrFenced
	}
	err = consumer(wire)
	clear(wire)
	return err
}
