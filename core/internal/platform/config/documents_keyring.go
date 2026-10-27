package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const documentsKEKKeyringFileKey = "HARPY_DOCUMENTS_KEK_KEYRING_FILE"

var errInvalidDocumentsKEKKeyring = errors.New("Documents KEK keyring is invalid")

type DocumentsKEK struct {
	id       string
	material [32]byte
	current  bool
}

type DocumentsKEKKeyring struct {
	releases []DocumentsKEK
	valid    bool
}

func loadDocumentsKEKKeyring(values environmentValues) (DocumentsKEKKeyring, error) {
	ring, err := loadFCARTKEKKeyringAt(values, documentsKEKKeyringFileKey,
		"documents-kek-keyring.v1")
	if err != nil {
		return DocumentsKEKKeyring{}, err
	}
	current, err := ring.Current()
	if err != nil {
		return DocumentsKEKKeyring{}, configError(documentsKEKKeyringFileKey,
			"must satisfy documents-kek-keyring.v1")
	}
	retained, err := ring.Retained()
	if err != nil {
		return DocumentsKEKKeyring{}, err
	}
	releases := make([]DocumentsKEK, 0, 1+len(retained))
	if release, ok := documentsRelease(current, true); ok {
		releases = append(releases, release)
	} else {
		return DocumentsKEKKeyring{}, invalidDocumentsKeyring()
	}
	for _, key := range retained {
		release, ok := documentsRelease(key, false)
		if !ok {
			return DocumentsKEKKeyring{}, invalidDocumentsKeyring()
		}
		releases = append(releases, release)
	}
	return DocumentsKEKKeyring{releases: releases, valid: true}, nil
}

func documentsRelease(key FCARTKEK, current bool) (DocumentsKEK, bool) {
	id, idErr := key.ID()
	material, keyErr := key.AES256Key()
	if idErr != nil || keyErr != nil || !validDocumentsKeyID(id) {
		return DocumentsKEK{}, false
	}
	return DocumentsKEK{id: id, material: material, current: current}, true
}

func (keyring DocumentsKEKKeyring) Validate() error {
	if !keyring.valid || len(keyring.releases) < 1 || len(keyring.releases) > 8 {
		return errInvalidDocumentsKEKKeyring
	}
	current := 0
	seen := make(map[string]struct{}, len(keyring.releases))
	for _, release := range keyring.releases {
		if !validDocumentsKeyID(release.id) || release.material == [32]byte{} {
			return errInvalidDocumentsKEKKeyring
		}
		if _, found := seen[release.id]; found {
			return errInvalidDocumentsKEKKeyring
		}
		seen[release.id] = struct{}{}
		if release.current {
			current++
		}
	}
	if current != 1 {
		return errInvalidDocumentsKEKKeyring
	}
	return nil
}

func (keyring DocumentsKEKKeyring) Releases() ([]DocumentsKEK, error) {
	if err := keyring.Validate(); err != nil {
		return nil, err
	}
	return append([]DocumentsKEK(nil), keyring.releases...), nil
}

func (key DocumentsKEK) Values() (string, [32]byte, bool, error) {
	if !validDocumentsKeyID(key.id) || key.material == [32]byte{} {
		return "", [32]byte{}, false, errInvalidDocumentsKEKKeyring
	}
	return key.id, key.material, key.current, nil
}

func validDocumentsKeyID(value string) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for _, character := range value {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '-' &&
			character != '_' && character != '.' {
			return false
		}
	}
	return true
}

func invalidDocumentsKeyring() error {
	return configError(documentsKEKKeyringFileKey,
		"must satisfy documents-kek-keyring.v1")
}

func (DocumentsKEKKeyring) String() string   { return "[REDACTED DOCUMENTS KEK KEYRING]" }
func (DocumentsKEKKeyring) GoString() string { return "[REDACTED DOCUMENTS KEK KEYRING]" }
func (DocumentsKEKKeyring) LogValue() slog.Value {
	return slog.StringValue("[REDACTED DOCUMENTS KEK KEYRING]")
}
func (DocumentsKEKKeyring) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED DOCUMENTS KEK KEYRING]")
}
