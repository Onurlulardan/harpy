package postrestore

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"net/url"
	"slices"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	store     Store
	security  Security
	directory DirectoryKey
	origins   []string
	rpID      string
	clock     Clock
	newUUID   UUIDGenerator
}

func NewService(
	store Store, security Security, directory DirectoryKey,
	origins []string, clock Clock, newUUID UUIDGenerator,
) (*Service, error) {
	if store == nil || directory.ID == "" || directory.Material == [32]byte{} ||
		security.Challenges == nil || security.Assertions == nil ||
		security.Passwords == nil || len(origins) == 0 || clock == nil || newUUID == nil {
		return nil, ErrUnavailable
	}
	parsed, err := url.Parse(origins[0])
	if err != nil || parsed.Hostname() == "" {
		return nil, ErrUnavailable
	}
	return &Service{store: store, security: security, directory: directory,
		origins: append([]string(nil), origins...), rpID: parsed.Hostname(),
		clock: clock, newUUID: newUUID}, nil
}

func (service *Service) Admission() Store { return service.store }

func (service *Service) validOrigin(origin string) bool {
	return slices.Contains(service.origins, origin)
}

func (service *Service) directoryHMAC(normalized string) [32]byte {
	mac := hmac.New(sha256.New, service.directory.Material[:])
	_, _ = mac.Write([]byte(normalized))
	var result [32]byte
	copy(result[:], mac.Sum(nil))
	return result
}

func newV7() (uuid.UUID, error) { return uuid.NewV7() }

func SystemClock() time.Time { return time.Now().UTC() }

func SystemUUID() (uuid.UUID, error) { return newV7() }

func mapStoreError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrUnavailable, err)
}
