package main

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"
)

const (
	listenKey    = "TELEMETRY_ARCHIVE_LISTEN_ADDR"
	certKey      = "TELEMETRY_ARCHIVE_TLS_CERT_FILE"
	privateKey   = "TELEMETRY_ARCHIVE_TLS_KEY_FILE"
	directoryKey = "TELEMETRY_ARCHIVE_DIRECTORY"
)

type runtimeConfig struct {
	listen    string
	cert      string
	key       string
	directory string
}

func loadConfig(lookup func(string) (string, bool)) (runtimeConfig, error) {
	config := runtimeConfig{}
	var err error
	if config.listen, err = required(lookup, listenKey); err != nil {
		return runtimeConfig{}, err
	}
	if _, _, err = net.SplitHostPort(config.listen); err != nil {
		return runtimeConfig{}, fmt.Errorf("%s is invalid", listenKey)
	}
	if config.cert, err = regularFile(lookup, certKey); err != nil {
		return runtimeConfig{}, err
	}
	if config.key, err = regularFile(lookup, privateKey); err != nil {
		return runtimeConfig{}, err
	}
	if _, err = tls.LoadX509KeyPair(config.cert, config.key); err != nil {
		return runtimeConfig{}, fmt.Errorf("telemetry TLS key pair is invalid")
	}
	if config.directory, err = required(lookup, directoryKey); err != nil {
		return runtimeConfig{}, err
	}
	info, err := os.Lstat(config.directory)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() ||
		!filepath.IsAbs(config.directory) {
		return runtimeConfig{}, fmt.Errorf("%s is invalid", directoryKey)
	}
	return config, nil
}

func required(lookup func(string) (string, bool), key string) (string, error) {
	value, found := lookup(key)
	if !found || value == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return value, nil
}

func regularFile(lookup func(string) (string, bool), key string) (string, error) {
	value, err := required(lookup, key)
	if err != nil {
		return "", err
	}
	info, statErr := os.Lstat(value)
	if statErr != nil || !filepath.IsAbs(value) || info.Mode()&os.ModeSymlink != 0 ||
		!info.Mode().IsRegular() {
		return "", fmt.Errorf("%s is invalid", key)
	}
	return value, nil
}
