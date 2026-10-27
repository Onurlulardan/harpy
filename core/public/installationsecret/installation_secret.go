package installationsecret

import internal "github.com/Onurlulardan/harpy/core/internal/platform/installationsecret"

var (
	ErrInvalid     = internal.ErrInvalid
	ErrExists      = internal.ErrExists
	ErrFenced      = internal.ErrFenced
	ErrInterrupted = internal.ErrInterrupted
)

type TemporaryPassword = internal.TemporaryPassword

func GenerateTemporaryPassword() (*TemporaryPassword, error) {
	return internal.GenerateTemporaryPassword()
}

func Publish(path string, wire []byte) error {
	return internal.Publish(path, wire)
}

func Read(path string, consumer func([]byte) error) error {
	return internal.Read(path, consumer)
}

func Remove(path string) error {
	return internal.Remove(path)
}
