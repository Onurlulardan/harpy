package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

var errInvalidFormsImportSource = errors.New("Forms import source artifact config is invalid")

type FormsImportSourceArtifacts struct {
	storage FormsImportSourceS3
	keyring FormsImportSourceKEKKeyring
	valid   bool
}

func loadFormsImportSourceArtifacts(values environmentValues, environment Environment,
	export FormsExportArtifacts) (FormsImportSourceArtifacts, error) {
	storage, err := loadFormsImportSourceS3(values, environment)
	if err != nil {
		return FormsImportSourceArtifacts{}, err
	}
	keyring, err := loadFormsImportSourceKEKKeyring(values)
	if err != nil {
		return FormsImportSourceArtifacts{}, err
	}
	if err := validateFormsArtifactKeySeparation(export.keyring, keyring); err != nil {
		return FormsImportSourceArtifacts{}, err
	}
	return FormsImportSourceArtifacts{storage: storage, keyring: keyring, valid: true}, nil
}

func (settings FormsImportSourceArtifacts) Validate() error {
	if !settings.valid || settings.storage.Validate() != nil || settings.keyring.Validate() != nil {
		return errInvalidFormsImportSource
	}
	return nil
}

func (settings FormsImportSourceArtifacts) Storage() (FormsImportSourceS3, error) {
	if err := settings.Validate(); err != nil {
		return FormsImportSourceS3{}, err
	}
	return settings.storage, nil
}

func (settings FormsImportSourceArtifacts) FCARTKeyring() (
	FormsImportSourceKEKKeyring, error,
) {
	if err := settings.Validate(); err != nil {
		return FormsImportSourceKEKKeyring{}, err
	}
	return settings.keyring, nil
}

func (FormsImportSourceArtifacts) String() string { return "[REDACTED FORMS IMPORT SOURCE CONFIG]" }
func (FormsImportSourceArtifacts) GoString() string {
	return "[REDACTED FORMS IMPORT SOURCE CONFIG]"
}
func (FormsImportSourceArtifacts) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FORMS IMPORT SOURCE CONFIG]")
}
func (FormsImportSourceArtifacts) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FORMS IMPORT SOURCE CONFIG]")
}
