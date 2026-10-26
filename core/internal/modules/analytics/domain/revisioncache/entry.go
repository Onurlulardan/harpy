package revisioncache

import (
	"bytes"
	"time"
)

const MaximumTTL = 300 * time.Second

type Entry struct {
	key       [32]byte
	result    []byte
	createdAt time.Time
	expiresAt time.Time
}

func NewEntry(binding Binding, result []byte, createdAt, expiresAt time.Time) (Entry, error) {
	key, err := binding.Key()
	if err != nil || len(result) == 0 || len(result) > 1<<20 || createdAt.IsZero() ||
		expiresAt.IsZero() || !expiresAt.After(createdAt) || expiresAt.Sub(createdAt) > MaximumTTL {
		return Entry{}, ErrInvalid
	}
	return Entry{key: key, result: bytes.Clone(result), createdAt: createdAt,
		expiresAt: expiresAt}, nil
}

func (entry Entry) Hit(binding Binding, now time.Time) ([]byte, bool) {
	key, err := binding.Key()
	if err != nil || now.IsZero() || now.Before(entry.createdAt) || !now.Before(entry.expiresAt) ||
		!bytes.Equal(key[:], entry.key[:]) {
		return nil, false
	}
	return bytes.Clone(entry.result), true
}
