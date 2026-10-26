package installationsecret

import (
	"errors"
	"os"
	"path/filepath"
)

func Publish(path string, wire []byte) (err error) {
	if !validPasswordWire(wire) {
		return ErrInvalid
	}
	target, err := validatedTarget(path, false)
	if err != nil {
		return err
	}
	staging, err := stagingPath(target)
	if err != nil {
		return err
	}
	file, err := createOwnerOnly(staging)
	if err != nil {
		return classifyCreateError(err)
	}
	published := false
	defer func() {
		if file != nil {
			_ = file.Close()
		}
		if !published {
			_ = os.Remove(staging)
		}
	}()
	if _, err = file.Write(wire); err == nil {
		err = file.Sync()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	file = nil
	if err != nil {
		return ErrInterrupted
	}
	if err = publishNoReplace(staging, target); err != nil {
		return err
	}
	published = true
	if err = syncDirectory(filepath.Dir(target)); err != nil {
		return ErrInterrupted
	}
	return nil
}

func classifyCreateError(err error) error {
	if errors.Is(err, os.ErrExist) {
		return ErrExists
	}
	return ErrFenced
}
