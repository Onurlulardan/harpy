package main

import (
	"crypto/rand"
	"errors"
	"regexp"
	"strings"
)

const (
	accessAlphabet  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	secretAlphabet  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	accessKeyLength = 20
	secretKeyLength = 40
	passwordLength  = 32
)

func fillSecrets(in answers, prev map[string]string, rotate bool) (answers, error) {
	var err error
	in.PostgresPassword, err = reuseOrGenerate(prev["POSTGRES_PASSWORD"], rotate, secretAlphabet, passwordLength)
	if err != nil {
		return in, err
	}
	in.RabbitPassword, err = reuseOrGenerate(prev["RABBITMQ_PASSWORD"], rotate, secretAlphabet, passwordLength)
	if err != nil {
		return in, err
	}
	in.RustFSAccessKey, err = reuseAccessKey(prev["RUSTFS_ACCESS_KEY"], rotate)
	if err != nil {
		return in, err
	}
	in.RustFSSecretKey, err = reuseOrGenerate(prev["RUSTFS_SECRET_KEY"], rotate, secretAlphabet, secretKeyLength)
	if err != nil {
		return in, err
	}
	if in.S3Provider == "rustfs" {
		return in.normalized(), nil
	}
	if in.S3AccessKey == "" {
		in.S3AccessKey = prev["S3_ACCESS_KEY"]
	}
	if in.S3SecretKey == "" {
		in.S3SecretKey = prev["S3_SECRET_KEY"]
	}
	return in.normalized(), nil
}

func reuseAccessKey(existing string, rotate bool) (string, error) {
	if !rotate && existing != "" {
		if err := validateAccessKey(existing); err != nil {
			return "", errors.New("stored access key cannot be kept; rerun with --rotate-secrets")
		}
		return existing, nil
	}
	for range 5 {
		key, err := randomString(accessAlphabet, accessKeyLength)
		if err != nil {
			return "", err
		}
		if validateAccessKey(key) == nil {
			return key, nil
		}
	}
	return "", errors.New("could not generate an access key")
}

func reuseOrGenerate(existing string, rotate bool, alphabet string, length int) (string, error) {
	if !rotate && existing != "" && existing != "rustfsadmin" && !strings.ContainsAny(existing, "\r\n") {
		return existing, nil
	}
	return randomString(alphabet, length)
}

func validateAccessKey(key string) error {
	if strings.EqualFold(key, "rustfsadmin") || !accessKeyPattern.MatchString(key) {
		return errors.New("access key must be A-Z and 0-9 and must not be rustfsadmin")
	}
	return nil
}

func randomString(alphabet string, length int) (string, error) {
	if length <= 0 || alphabet == "" {
		return "", errors.New("invalid random string length")
	}
	limit := 256 - (256 % len(alphabet))
	out := make([]byte, length)
	buf := make([]byte, length)
	filled := 0
	for filled < length {
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, item := range buf {
			if int(item) >= limit {
				continue
			}
			out[filled] = alphabet[int(item)%len(alphabet)]
			filled++
			if filled == length {
				break
			}
		}
	}
	return string(out), nil
}

var accessKeyPattern = regexp.MustCompile(`^[A-Z0-9]+$`)
