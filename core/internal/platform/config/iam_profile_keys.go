package config

const (
	iamTOTPKMSKeyIDEnvironmentKey     = "HARPY_IAM_TOTP_KMS_KEY_ID"
	iamDeliveryKMSKeyIDEnvironmentKey = "HARPY_IAM_DELIVERY_KMS_KEY_ID"
)

func iamAPIKeys() []string {
	keys := iamKeyringKeys(
		iamAccessKeyringEnvironmentKey,
		iamRefreshKeyringEnvironmentKey,
		iamCSRFKeyringEnvironmentKey,
		iamDirectoryKeyringEnvironmentKey,
		iamThrottleKeyringEnvironmentKey,
		iamSessionKeyringEnvironmentKey,
		iamOneTimeKeyringEnvironmentKey,
		iamRecoveryKeyringEnvironmentKey,
	)
	keys = append(keys,
		iamAccessIssuerEnvironmentKey,
		iamAccessAudienceEnvironmentKey,
		iamAllowedOriginsEnvironmentKey,
		iamPasswordBlocklistsEnvironmentKey,
		iamPasswordBlocklistsEnvironmentKey+"_FILE",
		iamTOTPKMSKeyIDEnvironmentKey,
		iamDeliveryKMSKeyIDEnvironmentKey,
	)
	return append(keys, iamActorBridgeKeys()...)
}

func iamWorkerKeys() []string {
	keys := iamKeyringKeys(
		iamDirectoryKeyringEnvironmentKey,
		iamOneTimeKeyringEnvironmentKey,
	)
	return append(keys, iamDeliveryKMSKeyIDEnvironmentKey)
}

func iamKeyringKeys(names ...string) []string {
	keys := make([]string, 0, len(names)*2)
	for _, name := range names {
		keys = append(keys, name, name+"_FILE")
	}
	return keys
}
