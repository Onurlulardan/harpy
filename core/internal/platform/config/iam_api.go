package config

import (
	"errors"
	"time"
)

var errInvalidIAMAPIConfig = errors.New("IAM API config is invalid")

type IAMAPI struct {
	iamKeyRedaction
	http           IAMHTTPProfile
	access         IAMAccessTokenKeyring
	refresh        IAMRefreshKeyring
	csrf           IAMCSRFKeyring
	directory      IAMDirectoryKeyring
	throttle       IAMThrottleKeyring
	sessionNetwork IAMSessionNetworkKeyring
	oneTime        IAMOneTimeKeyring
	recovery       IAMRecoveryKeyring
	passwords      IAMPasswordBlocklists
	totpKMS        IAMKMSKeyID
	deliveryKMS    IAMKMSKeyID
	actorBridge    IAMActorBridge
	valid          bool
}

func loadIAMAPI(
	values environmentValues,
	now time.Time,
	environment Environment,
) (IAMAPI, error) {
	httpProfile, err := loadIAMHTTPProfile(values, environment)
	if err != nil {
		return IAMAPI{}, err
	}
	access, err := loadIAMAccessKeyring(values, now)
	if err != nil {
		return IAMAPI{}, err
	}
	rawRings, err := loadIAMAPIHMACKeyrings(values, now)
	if err != nil {
		return IAMAPI{}, err
	}
	passwords, err := loadIAMPasswordBlocklists(values)
	if err != nil {
		return IAMAPI{}, err
	}
	totpKMS, err := loadIAMKMSKeyID(values, iamTOTPKMSKeyIDEnvironmentKey, environment)
	if err != nil {
		return IAMAPI{}, err
	}
	deliveryKMS, err := loadIAMKMSKeyID(values, iamDeliveryKMSKeyIDEnvironmentKey, environment)
	if err != nil {
		return IAMAPI{}, err
	}
	actorBridge, err := loadIAMActorBridge(values)
	if err != nil {
		return IAMAPI{}, err
	}
	config := newIAMAPI(httpProfile, access, rawRings, passwords, totpKMS, deliveryKMS,
		actorBridge)
	if err := config.Validate(); err != nil {
		return IAMAPI{}, configError(
			"HARPY_IAM_*", "must satisfy iam-key-purpose-isolation.v1",
		)
	}
	return config, nil
}

func newIAMAPI(
	httpProfile IAMHTTPProfile,
	access IAMAccessTokenKeyring,
	rings [7]iamHMACKeyring,
	passwords IAMPasswordBlocklists,
	totpKMS IAMKMSKeyID,
	deliveryKMS IAMKMSKeyID,
	actorBridge IAMActorBridge,
) IAMAPI {
	return IAMAPI{
		http:    httpProfile,
		access:  access,
		refresh: IAMRefreshKeyring{rings[0]}, csrf: IAMCSRFKeyring{rings[1]},
		directory: IAMDirectoryKeyring{rings[2]}, throttle: IAMThrottleKeyring{rings[3]},
		sessionNetwork: IAMSessionNetworkKeyring{rings[4]},
		oneTime:        IAMOneTimeKeyring{rings[5]}, recovery: IAMRecoveryKeyring{rings[6]},
		passwords: passwords,
		totpKMS:   totpKMS, deliveryKMS: deliveryKMS,
		actorBridge: actorBridge, valid: true,
	}
}

func (config IAMAPI) Validate() error {
	if !config.valid || config.http.Validate() != nil || config.access.Validate() != nil ||
		config.passwords.Validate() != nil ||
		config.totpKMS.Validate() != nil || config.deliveryKMS.Validate() != nil ||
		config.actorBridge.Validate() != nil {
		return errInvalidIAMAPIConfig
	}
	rings := config.hmacKeyrings()
	for index, ring := range rings {
		if ring.Validate() != nil || ring.purpose != iamAPIPurposes()[index] {
			return errInvalidIAMAPIConfig
		}
	}
	if !iamPurposeKeysUnique(
		config.access, rings, config.actorBridge,
		config.totpKMS.value, config.deliveryKMS.value,
	) || config.totpKMS.sameCryptoKey(config.deliveryKMS) {
		return errInvalidIAMAPIConfig
	}
	return nil
}

func (config IAMAPI) hmacKeyrings() []iamHMACKeyring {
	return []iamHMACKeyring{
		config.refresh.iamHMACKeyring, config.csrf.iamHMACKeyring,
		config.directory.iamHMACKeyring, config.throttle.iamHMACKeyring,
		config.sessionNetwork.iamHMACKeyring, config.oneTime.iamHMACKeyring,
		config.recovery.iamHMACKeyring,
	}
}
