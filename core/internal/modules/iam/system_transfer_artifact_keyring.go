package iam

import (
	"fmt"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func newSystemTransferArtifactKeyring(
	settings config.IAMTransferArtifactSettings,
) (*fcartv1.EnvelopeKeyring, error) {
	configured, err := settings.FCARTKeyring()
	if err != nil {
		return nil, err
	}
	currentConfig, err := configured.Current()
	if err != nil {
		return nil, err
	}
	current, err := systemTransferArtifactKEK(currentConfig)
	if err != nil {
		return nil, err
	}
	retainedConfig, err := configured.Retained()
	if err != nil {
		return nil, err
	}
	retained := make([]fcartv1.KEK, 0, len(retainedConfig))
	for _, configuredKey := range retainedConfig {
		key, keyErr := systemTransferArtifactKEK(configuredKey)
		if keyErr != nil {
			return nil, keyErr
		}
		retained = append(retained, key)
	}
	keyring, err := fcartv1.NewEnvelopeKeyring(current, retained, nil)
	if err != nil {
		return nil, fmt.Errorf("build IAM transfer FCART keyring: %w", err)
	}
	return keyring, nil
}

func systemTransferArtifactKEK(value config.FCARTKEK) (fcartv1.KEK, error) {
	id, err := value.ID()
	if err != nil {
		return fcartv1.KEK{}, err
	}
	material, err := value.AES256Key()
	if err != nil {
		return fcartv1.KEK{}, err
	}
	defer clear(material[:])
	key, err := fcartv1.NewKEK(id, material[:])
	if err != nil {
		return fcartv1.KEK{}, fmt.Errorf("build IAM transfer FCART KEK: %w", err)
	}
	return key, nil
}
