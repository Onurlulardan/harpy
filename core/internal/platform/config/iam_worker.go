package config

import (
	"errors"
	"time"
)

var errInvalidIAMWorkerConfig = errors.New("IAM worker config is invalid")

type IAMWorker struct {
	iamKeyRedaction
	directory      IAMDirectoryKeyring
	oneTime        IAMOneTimeKeyring
	deliveryKMS    IAMKMSKeyID
	invitationSMTP IAMInvitationSMTP
	valid          bool
}

func loadIAMWorker(values environmentValues, environment Environment, now time.Time) (IAMWorker, error) {
	ring, err := loadIAMHMACKeyring(
		values, iamDirectoryKeyringEnvironmentKey, iamPurposeDirectory, now,
	)
	if err != nil {
		return IAMWorker{}, err
	}
	oneTimeRing, err := loadIAMHMACKeyring(
		values, iamOneTimeKeyringEnvironmentKey, iamPurposeOneTime, now,
	)
	if err != nil {
		return IAMWorker{}, err
	}
	deliveryKMS, err := loadIAMKMSKeyID(values, iamDeliveryKMSKeyIDEnvironmentKey)
	if err != nil {
		return IAMWorker{}, err
	}
	smtp, err := loadIAMInvitationSMTP(values, environment)
	if err != nil {
		return IAMWorker{}, err
	}
	config := IAMWorker{
		directory: IAMDirectoryKeyring{ring}, oneTime: IAMOneTimeKeyring{oneTimeRing},
		deliveryKMS: deliveryKMS, invitationSMTP: smtp, valid: true,
	}
	if err := config.Validate(); err != nil {
		return IAMWorker{}, configError(
			"HARPY_IAM_*", "must satisfy IAM worker key-purpose isolation",
		)
	}
	return config, nil
}

func (config IAMWorker) Validate() error {
	if !config.valid || config.directory.Validate() != nil ||
		config.directory.purpose != iamPurposeDirectory ||
		config.oneTime.Validate() != nil || config.oneTime.purpose != iamPurposeOneTime ||
		config.deliveryKMS.Validate() != nil ||
		config.invitationSMTP.Validate() != nil {
		return errInvalidIAMWorkerConfig
	}
	if !iamWorkerKeysUnique(config.directory.iamHMACKeyring,
		config.oneTime.iamHMACKeyring, config.deliveryKMS.value) {
		return errInvalidIAMWorkerConfig
	}
	return nil
}

func iamWorkerKeysUnique(directory, oneTime iamHMACKeyring, deliveryKMS string) bool {
	seenIDs := make(map[string]struct{})
	seenMaterials := make([][32]byte, 0, 2)
	for _, ring := range []iamHMACKeyring{directory, oneTime} {
		if !addIAMPurposeMaterials(ring.keyMaterials(), seenIDs, &seenMaterials) {
			return false
		}
	}
	if _, duplicate := seenIDs[deliveryKMS]; duplicate || deliveryKMS == "" {
		return false
	}
	return true
}

func (config IAMWorker) OneTimeTokenKeyring() (IAMOneTimeKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMOneTimeKeyring{}, err
	}
	return config.oneTime, nil
}

func (config IAMWorker) DeliveryKMSKeyID() (IAMKMSKeyID, error) {
	if err := config.Validate(); err != nil {
		return IAMKMSKeyID{}, err
	}
	return config.deliveryKMS, nil
}

func (config IAMWorker) InvitationSMTP() (IAMInvitationSMTP, error) {
	if err := config.Validate(); err != nil {
		return IAMInvitationSMTP{}, err
	}
	return config.invitationSMTP, nil
}

func (config IAMWorker) LoginDirectoryKeyring() (IAMDirectoryKeyring, error) {
	if err := config.Validate(); err != nil {
		return IAMDirectoryKeyring{}, err
	}
	return config.directory, nil
}
