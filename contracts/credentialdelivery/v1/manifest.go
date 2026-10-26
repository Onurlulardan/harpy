package credentialdeliveryv1

import (
	"crypto/sha256"
	"encoding/hex"
)

type Command string

const (
	SupersedeTemporaryCredential Command = "iam.supersede_installation_temporary_credential"
	RecordDeliveryAttempt        Command = "installation.record_credential_delivery_attempt"
)

type Cause string

const (
	CauseClaimLeaseExpired Cause = "claim_lease_expired"
	CauseOutcomeUnknown    Cause = "outcome_unknown"
)

func Manifest() []Command {
	return []Command{SupersedeTemporaryCredential, RecordDeliveryAttempt}
}

func ManifestSHA256() string {
	digest := sha256.Sum256([]byte(
		string(SupersedeTemporaryCredential) + "\x00" + string(RecordDeliveryAttempt),
	))
	return hex.EncodeToString(digest[:])
}
