package config

import "time"

func loadIAMAPIHMACKeyrings(
	values environmentValues,
	now time.Time,
) ([7]iamHMACKeyring, error) {
	keys := []string{
		iamRefreshKeyringEnvironmentKey, iamCSRFKeyringEnvironmentKey,
		iamDirectoryKeyringEnvironmentKey, iamThrottleKeyringEnvironmentKey,
		iamSessionKeyringEnvironmentKey, iamOneTimeKeyringEnvironmentKey,
		iamRecoveryKeyringEnvironmentKey,
	}
	purposes := iamAPIPurposes()
	var rings [7]iamHMACKeyring
	for index, key := range keys {
		ring, err := loadIAMHMACKeyring(values, key, purposes[index], now)
		if err != nil {
			return [7]iamHMACKeyring{}, err
		}
		rings[index] = ring
	}
	return rings, nil
}

func iamAPIPurposes() [7]iamHMACPurpose {
	return [7]iamHMACPurpose{
		iamPurposeRefresh, iamPurposeCSRF, iamPurposeDirectory,
		iamPurposeThrottle, iamPurposeSessionNetwork, iamPurposeOneTime,
		iamPurposeRecovery,
	}
}
