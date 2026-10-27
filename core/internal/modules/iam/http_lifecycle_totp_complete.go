package iam

import (
	"encoding/json"
	"net/http"
)

type lifecycleTOTPCompleteDocument struct {
	ChallengeID               string          `json:"challenge_id"`
	ChallengeKeyID            string          `json:"challenge_key_id"`
	Challenge                 json.RawMessage `json:"challenge"`
	Code                      json.RawMessage `json:"code"`
	ExpectedChallengeRevision int64           `json:"expected_challenge_revision"`
}

func (handler *LifecycleHTTP) completeTOTP(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	authenticatorID, authValid := lifecyclePathUUID(r, "authenticator_id")
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	document, err := decodeIdentityDocument[lifecycleTOTPCompleteDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	challengeID, challengeValid := decodeUUID(document.ChallengeID)
	challenge, challengeErr := decodeLoginPassword(document.Challenge)
	code, codeErr := decodeLoginPassword(document.Code)
	clear(document.Challenge)
	clear(document.Code)
	if !authValid || !keyValid || !revisionValid || !challengeValid ||
		document.ChallengeKeyID == "" || document.ExpectedChallengeRevision < 1 ||
		challengeErr != nil || codeErr != nil {
		clear(challenge)
		clear(code)
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	defer clear(challenge)
	defer clear(code)
	err = handler.application.CompleteTOTP(r.Context(), claims, LifecycleTOTPCompleteInput{
		IdempotencyKey: key, ChallengeKeyID: document.ChallengeKeyID,
		AuthenticatorID: authenticatorID, ChallengeID: challengeID,
		ExpectedAuthenticatorRevision: revision,
		ExpectedChallengeRevision:     document.ExpectedChallengeRevision,
		Challenge:                     challenge, Code: code})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
