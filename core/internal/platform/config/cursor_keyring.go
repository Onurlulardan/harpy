package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const (
	cursorKeyringEnvironmentKey = "HARPY_REST_QUERY_CURSOR_KEYRING"
	cursorKeySize               = 32
	cursorRedacted              = "[REDACTED]"
)

type CursorKeyring struct {
	current  CursorKey
	previous CursorKey
	valid    bool
}

type CursorKey struct {
	kid      string
	material [cursorKeySize]byte
	valid    bool
}

var (
	errInvalidCursorKeyring = errors.New("REST Query cursor keyring is invalid")
	errInvalidCursorKey     = errors.New("REST Query cursor key is invalid")
)

func (k CursorKeyring) Validate() error {
	if !k.valid || k.current.validate() != nil || k.previous.validate() != nil ||
		k.current.kid == k.previous.kid {
		return errInvalidCursorKeyring
	}
	return nil
}

func (k CursorKeyring) Current() (CursorKey, error) {
	if err := k.Validate(); err != nil {
		return CursorKey{}, err
	}
	return k.current, nil
}

func (k CursorKeyring) Previous() (CursorKey, error) {
	if err := k.Validate(); err != nil {
		return CursorKey{}, err
	}
	return k.previous, nil
}

func (k CursorKey) KID() (string, error) {
	if err := k.validate(); err != nil {
		return "", err
	}
	return k.kid, nil
}

func (k CursorKey) AES256Key() ([cursorKeySize]byte, error) {
	if err := k.validate(); err != nil {
		return [cursorKeySize]byte{}, err
	}
	return k.material, nil
}

func (k CursorKey) validate() error {
	if !k.valid || !validCursorKID(k.kid) {
		return errInvalidCursorKey
	}
	return nil
}

func (CursorKeyring) String() string   { return cursorRedacted }
func (CursorKeyring) GoString() string { return cursorRedacted }
func (CursorKeyring) LogValue() slog.Value {
	return slog.StringValue(cursorRedacted)
}
func (CursorKeyring) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}
func (CursorKeyring) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, cursorRedacted)
}

func (CursorKey) String() string   { return cursorRedacted }
func (CursorKey) GoString() string { return cursorRedacted }
func (CursorKey) LogValue() slog.Value {
	return slog.StringValue(cursorRedacted)
}
func (CursorKey) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED]"`), nil
}
func (CursorKey) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, cursorRedacted)
}
