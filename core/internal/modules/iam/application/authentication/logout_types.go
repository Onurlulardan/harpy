package authentication

import "github.com/google/uuid"

type LogoutRequest struct {
	Presented PresentedRefresh
}

func (request LogoutRequest) validate() error {
	if nilInterface(request.Presented) {
		return ErrAuthenticationContract
	}
	return nil
}

type LogoutMutation struct {
	TokenID, FamilyID, SessionID, UserID          uuid.UUID
	ExpectedTokenRevision, ExpectedFamilyRevision int64
	ExpectedSessionRevision, ExpectedUserRevision int64
}

type LogoutOutcome string

const (
	LogoutRevoked  LogoutOutcome = "revoked"
	LogoutTerminal LogoutOutcome = "terminal"
)

type LogoutResult struct{ Outcome LogoutOutcome }

func (result LogoutResult) validate() error {
	if result.Outcome != LogoutRevoked && result.Outcome != LogoutTerminal {
		return ErrAuthenticationStorage
	}
	return nil
}

func (credential RefreshCredential) logoutMutation() LogoutMutation {
	return LogoutMutation{
		TokenID: credential.TokenID, FamilyID: credential.FamilyID,
		SessionID: credential.SessionID, UserID: credential.UserID,
		ExpectedTokenRevision:   credential.TokenRevision,
		ExpectedFamilyRevision:  credential.FamilyRevision,
		ExpectedSessionRevision: credential.SessionRevision,
		ExpectedUserRevision:    credential.UserRevision,
	}
}
