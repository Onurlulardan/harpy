package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRestoredSecretTreeKeepsFingerprints(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	secrets := filepath.Join(root, "deploy", "single-tenant", "secrets")
	first := secretFingerprints(t, root, "single-tenant")
	aside := filepath.Join(t.TempDir(), "secrets")
	copySecretTree(t, secrets, aside)
	if err := os.RemoveAll(secrets); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(secrets); !os.IsNotExist(err) {
		t.Fatal("working secrets tree was not removed")
	}
	copySecretTree(t, aside, secrets)
	assertSameFingerprints(t, first, secretFingerprints(t, root, "single-tenant"))
	runOK(t, root, []string{"--defaults"}, "")
	assertSameFingerprints(t, first, secretFingerprints(t, root, "single-tenant"))
}

func TestStayedAccessKeyringParsesUntilRotate(t *testing.T) {
	root := t.TempDir()
	runOK(t, root, []string{"--defaults"}, "")
	path := filepath.Join(root, "deploy", "single-tenant", "secrets", "runtime", "access-token-keyring.json")
	assertAccessKeyringParseable(t, path)
	before := fileFingerprint(t, path)
	runOK(t, root, []string{"--defaults"}, "")
	assertAccessKeyringParseable(t, path)
	if fileFingerprint(t, path) != before {
		t.Fatal("second setup changed the stayed access token keyring")
	}
	runOK(t, root, []string{"--defaults", "--rotate-secrets"}, "")
	assertAccessKeyringParseable(t, path)
	if fileFingerprint(t, path) == before {
		t.Fatal("rotate-secrets kept the access token keyring")
	}
}

func copySecretTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, 0o700)
		}
		if !entry.Type().IsRegular() {
			return errors.New("secret tree entry is not a regular file")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o600)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func assertAccessKeyringParseable(t *testing.T, path string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var doc accessKeyringDocument
	if err = dec.Decode(&doc); err != nil {
		t.Fatal("access token keyring is not parseable")
	}
	var extra struct{}
	if err = dec.Decode(&extra); err != io.EOF {
		t.Fatal("access token keyring has trailing data")
	}
	if doc.Version != 1 || doc.Active.KID == "" || doc.Active.Seed == "" {
		t.Fatal("access token keyring is missing active material")
	}
	if doc.RetiredPublic == nil {
		t.Fatal("access token keyring retired_public is missing")
	}
	seed, err := base64.RawURLEncoding.DecodeString(doc.Active.Seed)
	if err != nil || len(seed) != 32 {
		t.Fatal("access token keyring seed is not a 32-byte key")
	}
	activated, err := time.Parse(time.RFC3339, doc.Active.ActivatedAt)
	if err != nil {
		t.Fatal("access token keyring activated_at is not a timestamp")
	}
	until, err := time.Parse(time.RFC3339, doc.Active.SignUntil)
	if err != nil || !until.After(activated) {
		t.Fatal("access token keyring sign_until is not after activation")
	}
}

type accessKeyringDocument struct {
	Version int `json:"version"`
	Active  struct {
		KID         string `json:"kid"`
		Seed        string `json:"seed"`
		ActivatedAt string `json:"activated_at"`
		SignUntil   string `json:"sign_until"`
	} `json:"active"`
	RetiredPublic []json.RawMessage `json:"retired_public"`
}
