package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

var errInvalidFormsExportArtifacts = errors.New("Forms export artifact config is invalid")

type FormsExportArtifacts struct {
	storage FormsExportS3
	keyring FCARTKEKKeyring
	valid   bool
}

func loadFormsExportArtifacts(
	values environmentValues,
	environment Environment,
) (FormsExportArtifacts, error) {
	storage, err := loadFormsExportS3(values, environment)
	if err != nil {
		return FormsExportArtifacts{}, err
	}
	keyring, err := loadFCARTKEKKeyring(values)
	if err != nil {
		return FormsExportArtifacts{}, err
	}
	return FormsExportArtifacts{storage: storage, keyring: keyring, valid: true}, nil
}

func (settings FormsExportArtifacts) Validate() error {
	if !settings.valid || settings.storage.Validate() != nil || settings.keyring.Validate() != nil {
		return errInvalidFormsExportArtifacts
	}
	return nil
}

func (settings FormsExportArtifacts) Storage() (FormsExportS3, error) {
	if err := settings.Validate(); err != nil {
		return FormsExportS3{}, err
	}
	return settings.storage, nil
}

func (settings FormsExportArtifacts) FCARTKeyring() (FCARTKEKKeyring, error) {
	if err := settings.Validate(); err != nil {
		return FCARTKEKKeyring{}, err
	}
	return settings.keyring, nil
}

func (FormsExportArtifacts) String() string   { return "[REDACTED FORMS EXPORT ARTIFACT CONFIG]" }
func (FormsExportArtifacts) GoString() string { return "[REDACTED FORMS EXPORT ARTIFACT CONFIG]" }
func (FormsExportArtifacts) LogValue() slog.Value {
	return slog.StringValue("[REDACTED FORMS EXPORT ARTIFACT CONFIG]")
}
func (FormsExportArtifacts) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED FORMS EXPORT ARTIFACT CONFIG]")
}
