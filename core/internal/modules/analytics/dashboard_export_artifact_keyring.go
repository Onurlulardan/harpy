package analytics

import (
	"fmt"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func NewDashboardExportArtifactKeyring(
	artifacts config.AnalyticsDashboardExportArtifacts,
) (*fcartv1.EnvelopeKeyring, error) {
	configured, err := artifacts.FCARTKeyring()
	if err != nil {
		return nil, fmt.Errorf("load Analytics export keyring: %w", err)
	}
	currentConfig, err := configured.Current()
	if err != nil {
		return nil, fmt.Errorf("load current Analytics export KEK: %w", err)
	}
	current, err := dashboardExportKEK(currentConfig)
	if err != nil {
		return nil, err
	}
	retainedConfig, err := configured.Retained()
	if err != nil {
		return nil, fmt.Errorf("load retained Analytics export KEKs: %w", err)
	}
	retained := make([]fcartv1.KEK, 0, len(retainedConfig))
	for _, configuredKey := range retainedConfig {
		key, keyErr := dashboardExportKEK(configuredKey)
		if keyErr != nil {
			return nil, keyErr
		}
		retained = append(retained, key)
	}
	keys, err := fcartv1.NewEnvelopeKeyring(current, retained, nil)
	if err != nil {
		return nil, fmt.Errorf("build Analytics export keyring: %w", err)
	}
	return keys, nil
}

func dashboardExportKEK(configured config.FCARTKEK) (fcartv1.KEK, error) {
	id, err := configured.ID()
	if err != nil {
		return fcartv1.KEK{}, fmt.Errorf("load Analytics export KEK ID: %w", err)
	}
	material, err := configured.AES256Key()
	if err != nil {
		return fcartv1.KEK{}, fmt.Errorf("load Analytics export KEK material: %w", err)
	}
	defer clear(material[:])
	key, err := fcartv1.NewKEK(id, material[:])
	if err != nil {
		return fcartv1.KEK{}, fmt.Errorf("build Analytics export KEK: %w", err)
	}
	return key, nil
}
