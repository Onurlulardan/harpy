package authentication

import "github.com/google/uuid"

type ReauthenticationRejection struct {
	ChallengeID                   uuid.UUID
	ExpectedRevision              int64
	ExpectedAttemptCount          int16
	AuthenticatorID               uuid.UUID
	ExpectedAuthenticatorRevision int64
	QuarantineAuthenticator       bool
	Expire                        bool
}

type ReauthenticationRejected struct {
	State        string
	Revision     int64
	AttemptCount int16
}

func (rejected ReauthenticationRejected) validate(
	rejection ReauthenticationRejection,
) error {
	if rejected.Revision != rejection.ExpectedRevision+1 ||
		rejected.AttemptCount < rejection.ExpectedAttemptCount ||
		rejected.AttemptCount > 5 ||
		(rejected.State != "pending" && rejected.State != "failed" &&
			rejected.State != "expired") {
		return ErrAuthenticationStorage
	}
	if rejection.Expire && rejected.State != "expired" {
		return ErrAuthenticationStorage
	}
	return nil
}
