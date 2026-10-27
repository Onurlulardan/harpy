package config

import "time"

type iamHMACPurpose string

const (
	iamPurposeRefresh        iamHMACPurpose = "refresh_token_v1"
	iamPurposeCSRF           iamHMACPurpose = "csrf_v1"
	iamPurposeDirectory      iamHMACPurpose = "login_directory_v1"
	iamPurposeThrottle       iamHMACPurpose = "auth_throttle_v1"
	iamPurposeSessionNetwork iamHMACPurpose = "session_network_v1"
	iamPurposeOneTime        iamHMACPurpose = "one_time_token_v1"
	iamPurposeRecovery       iamHMACPurpose = "recovery_code_v1"

	iamAccessKeyringEnvironmentKey    = "HARPY_IAM_ACCESS_TOKEN_KEYRING"
	iamRefreshKeyringEnvironmentKey   = "HARPY_IAM_REFRESH_TOKEN_KEYRING"
	iamCSRFKeyringEnvironmentKey      = "HARPY_IAM_CSRF_KEYRING"
	iamDirectoryKeyringEnvironmentKey = "HARPY_IAM_LOGIN_DIRECTORY_KEYRING"
	iamThrottleKeyringEnvironmentKey  = "HARPY_IAM_AUTH_THROTTLE_KEYRING"
	iamSessionKeyringEnvironmentKey   = "HARPY_IAM_SESSION_NETWORK_KEYRING"
	iamOneTimeKeyringEnvironmentKey   = "HARPY_IAM_ONE_TIME_TOKEN_KEYRING"
	iamRecoveryKeyringEnvironmentKey  = "HARPY_IAM_RECOVERY_CODE_KEYRING"
)

const iamHMACRotation = 90 * 24 * time.Hour

func (purpose iamHMACPurpose) valid() bool {
	switch purpose {
	case iamPurposeRefresh, iamPurposeCSRF, iamPurposeDirectory,
		iamPurposeThrottle, iamPurposeSessionNetwork, iamPurposeOneTime,
		iamPurposeRecovery:
		return true
	default:
		return false
	}
}

func (purpose iamHMACPurpose) receiptGated() bool {
	return purpose == iamPurposeDirectory || purpose == iamPurposeRecovery
}

func (purpose iamHMACPurpose) minimumRetention() time.Duration {
	switch purpose {
	case iamPurposeRefresh, iamPurposeSessionNetwork:
		return 91 * 24 * time.Hour
	case iamPurposeCSRF:
		return 16 * time.Minute
	case iamPurposeThrottle:
		return 75 * time.Minute
	case iamPurposeOneTime:
		return 8 * 24 * time.Hour
	default:
		return 0
	}
}
