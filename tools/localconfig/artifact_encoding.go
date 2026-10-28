package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func randomBytes(size int) ([]byte, error) {
	if size < 1 {
		return nil, fmt.Errorf("random byte size must be positive")
	}
	for {
		value := make([]byte, size)
		if _, err := rand.Read(value); err != nil {
			return nil, err
		}
		if !allZero(value) {
			return value, nil
		}
	}
}

// randomBase64 values are consumed as positional CLI arguments (mc alias set,
// mc admin user add), env-file values, DSN userinfo and JSON fields. A leading
// dash would be parsed as a flag by every urfave/cli-style consumer, so the
// first character must never be '-'.
func randomBase64(size int) (string, error) {
	for {
		value, err := randomBytes(size)
		if err != nil {
			return "", err
		}
		encoded := base64.RawURLEncoding.EncodeToString(value)
		clear(value)
		if encoded[0] != '-' {
			return encoded, nil
		}
	}
}

func randomLowerID(prefix string) (string, error) {
	value, err := randomBytes(6)
	if err != nil {
		return "", err
	}
	id := prefix + hex.EncodeToString(value)
	clear(value)
	return id, nil
}

func randomAccessKeyID() (string, error) {
	value, err := randomBytes(10)
	if err != nil {
		return "", err
	}
	id := strings.ToUpper(hex.EncodeToString(value))
	clear(value)
	return id, nil
}

func marshalJSON(value any) ([]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("marshal JSON: %w", err)
	}
	return document, nil
}

func addJSON(set *artifactSet, name string, value any) error {
	document, err := marshalJSON(value)
	if err != nil {
		return err
	}
	return set.add(name, document)
}

func canonicalTimestamp(value time.Time) string {
	return value.UTC().Truncate(time.Second).Format(time.RFC3339)
}

func sha256Hex(value []byte) string {
	digest := sha256.Sum256(value)
	return hex.EncodeToString(digest[:])
}

func allZero(value []byte) bool {
	for _, item := range value {
		if item != 0 {
			return false
		}
	}
	return true
}
