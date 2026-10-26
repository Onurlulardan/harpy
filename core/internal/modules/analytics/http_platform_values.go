package analytics

import (
	"crypto/rand"
	"io"
	"time"

	"github.com/google/uuid"
)

type httpPlatformValues struct{}

func (httpPlatformValues) NewUUIDv7() (uuid.UUID, error) { return uuid.NewV7() }
func (httpPlatformValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}
func (httpPlatformValues) Read(target []byte) error {
	read, err := rand.Read(target)
	if err != nil || read != len(target) {
		if err != nil {
			return err
		}
		return io.ErrUnexpectedEOF
	}
	return nil
}
