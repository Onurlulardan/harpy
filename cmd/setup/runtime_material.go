package main

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"path/filepath"
	"time"
)

func runtimeMaterial(root, profile string) ([]string, error) {
	mode, err := deploymentMode(profile)
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(root, "deploy", profile, "secrets", "runtime")
	if err = os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	now := time.Now().UTC().Truncate(time.Second)
	lines, err := writeIdentityMaterial(dir, now)
	if err != nil {
		return nil, err
	}
	services, err := writeServiceMaterial(dir)
	if err != nil {
		return nil, err
	}
	authority, err := writeAuthorityMaterial(dir, mode)
	if err != nil {
		return nil, err
	}
	lines = append(lines, services...)
	return append(lines, authority...), nil
}

func writeRuntimeFile(dir, name string, data []byte) (string, error) {
	return writeMaterialFile(dir, name, data, !setupRotateSecrets)
}

func runtimeDirectory(parent, name string) (string, error) {
	path := filepath.Join(parent, name)
	if err := os.MkdirAll(path, 0o700); err != nil {
		return "", err
	}
	return path, nil
}

func runtimeLine(key, value string) string {
	return key + "=" + value
}

func markedMaterial(marker byte) ([32]byte, error) {
	var material [32]byte
	if _, err := rand.Read(material[:]); err != nil {
		return material, err
	}
	material[0] = marker
	material[31] = marker
	return material, nil
}

func rawKey(material []byte) string {
	return base64.RawURLEncoding.EncodeToString(material)
}

func iamKeyID(material [32]byte) string {
	return rawKey(material[:16])
}

func iamStamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}
