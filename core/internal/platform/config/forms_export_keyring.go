package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

const formsExportFCARTKeyringFileKey = "HARPY_FORMS_EXPORT_FCART_KEK_KEYRING_FILE"

var (
	errInvalidFCARTKEK        = errors.New("FCART KEK is invalid")
	errInvalidFCARTKEKKeyring = errors.New("FCART KEK keyring is invalid")
)

type FCARTKEK struct {
	id       string
	material [32]byte
	valid    bool
}

type FCARTKEKKeyring struct {
	current  FCARTKEK
	retained []FCARTKEK
	valid    bool
}

func (key FCARTKEK) Validate() error {
	if !key.valid || !validFCARTKeyID(key.id) || allZero(key.material[:]) {
		return errInvalidFCARTKEK
	}
	return nil
}

func (key FCARTKEK) ID() (string, error) {
	if err := key.Validate(); err != nil {
		return "", err
	}
	return key.id, nil
}

func (key FCARTKEK) AES256Key() ([32]byte, error) {
	if err := key.Validate(); err != nil {
		return [32]byte{}, err
	}
	return key.material, nil
}

func (keyring FCARTKEKKeyring) Validate() error {
	if !keyring.valid || keyring.current.Validate() != nil || len(keyring.retained) > 7 {
		return errInvalidFCARTKEKKeyring
	}
	seen := map[string]struct{}{keyring.current.id: {}}
	for _, key := range keyring.retained {
		if key.Validate() != nil {
			return errInvalidFCARTKEKKeyring
		}
		if _, duplicate := seen[key.id]; duplicate {
			return errInvalidFCARTKEKKeyring
		}
		seen[key.id] = struct{}{}
	}
	return nil
}

func (keyring FCARTKEKKeyring) Current() (FCARTKEK, error) {
	if err := keyring.Validate(); err != nil {
		return FCARTKEK{}, err
	}
	return keyring.current, nil
}

func (keyring FCARTKEKKeyring) Retained() ([]FCARTKEK, error) {
	if err := keyring.Validate(); err != nil {
		return nil, err
	}
	return append([]FCARTKEK(nil), keyring.retained...), nil
}

func (FCARTKEK) String() string   { return "[REDACTED FCART KEK]" }
func (FCARTKEK) GoString() string { return "[REDACTED FCART KEK]" }
func (FCARTKEK) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FCART KEK]")
}
func (FCARTKEK) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FCART KEK]")
}

func (FCARTKEKKeyring) String() string   { return "[REDACTED FCART KEK KEYRING]" }
func (FCARTKEKKeyring) GoString() string { return "[REDACTED FCART KEK KEYRING]" }
func (FCARTKEKKeyring) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FCART KEK KEYRING]")
}
func (FCARTKEKKeyring) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FCART KEK KEYRING]")
}

func validFCARTKeyID(value string) bool {
	if len(value) < 1 || len(value) > 128 || !alphaNumeric(value[0]) {
		return false
	}
	for index := 1; index < len(value); index++ {
		if !alphaNumeric(value[index]) && !containsByte("_.:-", value[index]) {
			return false
		}
	}
	return true
}

func alphaNumeric(value byte) bool {
	return value >= 'A' && value <= 'Z' || value >= 'a' && value <= 'z' ||
		value >= '0' && value <= '9'
}

func containsByte(values string, target byte) bool {
	for index := 0; index < len(values); index++ {
		if values[index] == target {
			return true
		}
	}
	return false
}

func allZero(value []byte) bool {
	var combined byte
	for _, item := range value {
		combined |= item
	}
	return combined == 0
}
