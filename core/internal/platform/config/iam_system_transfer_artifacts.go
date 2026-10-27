package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

var errInvalidIAMTransferArtifacts = errors.New("IAM system transfer artifacts are invalid")

type IAMTransferArtifactSettings struct {
	storage ArtifactS3
	keyring FCARTKEKKeyring
	valid   bool
}

type IAMSystemTransferArtifacts struct {
	export       IAMTransferArtifactSettings
	importSource IAMTransferArtifactSettings
	spool        AbsolutePath
	valid        bool
}

func loadIAMSystemTransferArtifacts(values environmentValues, environment Environment) (
	IAMSystemTransferArtifacts, error,
) {
	exportStorage, err := loadArtifactS3(values, environment, iamExportStorageKeys())
	if err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	importStorage, err := loadArtifactS3(values, environment, iamImportStorageKeys())
	if err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	exportKeys, err := loadFCARTKEKKeyringAt(values, iamExportKeyringKey,
		"iam-system-export-fcart-kek-keyring.v1")
	if err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	importKeys, err := loadFCARTKEKKeyringAt(values, iamImportKeyringKey,
		"iam-system-import-source-fcart-kek-keyring.v1")
	if err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	spool, err := loadAbsolutePath(values, iamTransferSpoolKey)
	if err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	settings := IAMSystemTransferArtifacts{
		export: IAMTransferArtifactSettings{storage: exportStorage,
			keyring: exportKeys, valid: true},
		importSource: IAMTransferArtifactSettings{storage: importStorage,
			keyring: importKeys, valid: true}, spool: spool, valid: true,
	}
	if err := validateIAMArtifactSeparation(settings); err != nil {
		return IAMSystemTransferArtifacts{}, err
	}
	return settings, nil
}

func (settings IAMTransferArtifactSettings) Validate() error {
	if !settings.valid || settings.storage.Validate() != nil ||
		settings.keyring.Validate() != nil {
		return errInvalidIAMTransferArtifacts
	}
	return nil
}

func (settings IAMTransferArtifactSettings) Storage() (ArtifactS3, error) {
	if err := settings.Validate(); err != nil {
		return ArtifactS3{}, err
	}
	return settings.storage, nil
}

func (settings IAMTransferArtifactSettings) FCARTKeyring() (FCARTKEKKeyring, error) {
	if err := settings.Validate(); err != nil {
		return FCARTKEKKeyring{}, err
	}
	return settings.keyring, nil
}

func (settings IAMSystemTransferArtifacts) Validate() error {
	if !settings.valid || settings.export.Validate() != nil ||
		settings.importSource.Validate() != nil || settings.spool.Value() == "" {
		return errInvalidIAMTransferArtifacts
	}
	return nil
}

func (settings IAMSystemTransferArtifacts) Export() (IAMTransferArtifactSettings, error) {
	if err := settings.Validate(); err != nil {
		return IAMTransferArtifactSettings{}, err
	}
	return settings.export, nil
}

func (settings IAMSystemTransferArtifacts) ImportSource() (
	IAMTransferArtifactSettings, error,
) {
	if err := settings.Validate(); err != nil {
		return IAMTransferArtifactSettings{}, err
	}
	return settings.importSource, nil
}

func (settings IAMSystemTransferArtifacts) SpoolDirectory() (AbsolutePath, error) {
	if err := settings.Validate(); err != nil {
		return AbsolutePath{}, err
	}
	return settings.spool, nil
}

func (IAMSystemTransferArtifacts) String() string   { return "[REDACTED IAM ARTIFACTS]" }
func (IAMSystemTransferArtifacts) GoString() string { return "[REDACTED IAM ARTIFACTS]" }
func (IAMSystemTransferArtifacts) LogValue() slog.Value {
	return slog.StringValue("[REDACTED IAM ARTIFACTS]")
}
func (IAMSystemTransferArtifacts) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED IAM ARTIFACTS]")
}
