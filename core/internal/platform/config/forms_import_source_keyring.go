package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const formsImportSourceFCARTKeyringFileKey = "HARPY_FORMS_IMPORT_SOURCE_FCART_KEK_KEYRING_FILE"

var (
	errInvalidFormsImportSourceKEK     = errors.New("Forms import source FCART KEK is invalid")
	errInvalidFormsImportSourceKeyring = errors.New(
		"Forms import source FCART KEK keyring is invalid")
)

type FormsImportSourceKEK struct {
	id       string
	material [32]byte
	valid    bool
}

type FormsImportSourceKEKKeyring struct {
	current  FormsImportSourceKEK
	retained []FormsImportSourceKEK
	valid    bool
}

func (key FormsImportSourceKEK) Validate() error {
	if !key.valid || !validFCARTKeyID(key.id) || allZero(key.material[:]) {
		return errInvalidFormsImportSourceKEK
	}
	return nil
}

func (key FormsImportSourceKEK) ID() (string, error) {
	if err := key.Validate(); err != nil {
		return "", err
	}
	return key.id, nil
}

func (key FormsImportSourceKEK) AES256Key() ([32]byte, error) {
	if err := key.Validate(); err != nil {
		return [32]byte{}, err
	}
	return key.material, nil
}

func (keyring FormsImportSourceKEKKeyring) Validate() error {
	if !keyring.valid || keyring.current.Validate() != nil || len(keyring.retained) > 7 {
		return errInvalidFormsImportSourceKeyring
	}
	seenIDs := map[string]struct{}{keyring.current.id: {}}
	seenMaterial := map[[32]byte]struct{}{keyring.current.material: {}}
	for _, key := range keyring.retained {
		_, duplicateID := seenIDs[key.id]
		_, duplicateMaterial := seenMaterial[key.material]
		if key.Validate() != nil || duplicateID || duplicateMaterial {
			return errInvalidFormsImportSourceKeyring
		}
		seenIDs[key.id], seenMaterial[key.material] = struct{}{}, struct{}{}
	}
	return nil
}

func (keyring FormsImportSourceKEKKeyring) Current() (FormsImportSourceKEK, error) {
	if err := keyring.Validate(); err != nil {
		return FormsImportSourceKEK{}, err
	}
	return keyring.current, nil
}

func (keyring FormsImportSourceKEKKeyring) Retained() ([]FormsImportSourceKEK, error) {
	if err := keyring.Validate(); err != nil {
		return nil, err
	}
	return append([]FormsImportSourceKEK(nil), keyring.retained...), nil
}

func (keyring FormsImportSourceKEKKeyring) all() []FormsImportSourceKEK {
	return append([]FormsImportSourceKEK{keyring.current}, keyring.retained...)
}

func (FormsImportSourceKEK) String() string   { return "[REDACTED FCART KEK]" }
func (FormsImportSourceKEK) GoString() string { return "[REDACTED FCART KEK]" }
func (FormsImportSourceKEK) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FCART KEK]")
}
func (FormsImportSourceKEK) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FCART KEK]")
}

func (FormsImportSourceKEKKeyring) String() string { return "[REDACTED FCART KEK KEYRING]" }
func (FormsImportSourceKEKKeyring) GoString() string {
	return "[REDACTED FCART KEK KEYRING]"
}
func (FormsImportSourceKEKKeyring) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FCART KEK KEYRING]")
}
func (FormsImportSourceKEKKeyring) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FCART KEK KEYRING]")
}
