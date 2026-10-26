package config

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
)

var errInvalidAnalyticsDashboardExportArtifacts = errors.New(
	"Analytics dashboard export artifacts are invalid",
)

type AnalyticsDashboardExportArtifacts struct {
	storage ArtifactS3
	keyring FCARTKEKKeyring
	valid   bool
}

func loadAnalyticsDashboardExportArtifacts(
	values environmentValues, environment Environment,
) (AnalyticsDashboardExportArtifacts, error) {
	storage, err := loadArtifactS3(
		values, environment, analyticsDashboardExportStorageKeys(),
	)
	if err != nil {
		return AnalyticsDashboardExportArtifacts{}, err
	}
	keyring, err := loadFCARTKEKKeyringAt(values, analyticsExportKeyringKey,
		"analytics-dashboard-export-fcart-kek-keyring.v1")
	if err != nil {
		return AnalyticsDashboardExportArtifacts{}, err
	}
	return AnalyticsDashboardExportArtifacts{
		storage: storage, keyring: keyring, valid: true,
	}, nil
}

func (settings AnalyticsDashboardExportArtifacts) Validate() error {
	if !settings.valid || settings.storage.Validate() != nil ||
		settings.keyring.Validate() != nil {
		return errInvalidAnalyticsDashboardExportArtifacts
	}
	return nil
}

func (settings AnalyticsDashboardExportArtifacts) Storage() (ArtifactS3, error) {
	if err := settings.Validate(); err != nil {
		return ArtifactS3{}, err
	}
	return settings.storage, nil
}

func (settings AnalyticsDashboardExportArtifacts) FCARTKeyring() (FCARTKEKKeyring, error) {
	if err := settings.Validate(); err != nil {
		return FCARTKEKKeyring{}, err
	}
	return settings.keyring, nil
}

func (AnalyticsDashboardExportArtifacts) String() string {
	return "[REDACTED ANALYTICS DASHBOARD EXPORT ARTIFACTS]"
}
func (AnalyticsDashboardExportArtifacts) GoString() string {
	return "[REDACTED ANALYTICS DASHBOARD EXPORT ARTIFACTS]"
}
func (AnalyticsDashboardExportArtifacts) LogValue() slog.Value {
	return slog.StringValue("[REDACTED ANALYTICS DASHBOARD EXPORT ARTIFACTS]")
}
func (AnalyticsDashboardExportArtifacts) Format(state fmt.State, _ rune) {
	_, _ = io.WriteString(state, "[REDACTED ANALYTICS DASHBOARD EXPORT ARTIFACTS]")
}
