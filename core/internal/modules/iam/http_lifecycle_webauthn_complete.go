package iam

import (
	"encoding/json"
	"net/http"
)

type lifecycleWebAuthnCompleteDocument struct {
	ChallengeID    string          `json:"challenge_id"`
	ChallengeKeyID string          `json:"challenge_key_id"`
	Challenge      json.RawMessage `json:"challenge"`
	Credential     json.RawMessage `json:"credential"`
}

func (handler *LifecycleHTTP) completeWebAuthn(w http.ResponseWriter, r *http.Request) {
	claims, valid := handler.claims(r)
	if !valid {
		handler.writeFailure(w, r, errLifecycleAuthentication)
		return
	}
	authenticatorID, authValid := lifecyclePathUUID(r, "authenticator_id")
	key, keyValid := identityIdempotencyKey(r)
	revision, revisionValid := identityExpectedRevision(r)
	document, err := decodeIdentityDocument[lifecycleWebAuthnCompleteDocument](w, r)
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	challengeID, challengeValid := decodeUUID(document.ChallengeID)
	challenge, challengeErr := decodeLoginPassword(document.Challenge)
	credential := append([]byte(nil), document.Credential...)
	clear(document.Challenge)
	clear(document.Credential)
	if !authValid || !keyValid || !revisionValid || !challengeValid ||
		document.ChallengeKeyID == "" || challengeErr != nil || len(credential) < 2 {
		clear(challenge)
		clear(credential)
		handler.writeFailure(w, r, ErrLifecycleHTTPInvalid)
		return
	}
	defer clear(challenge)
	defer clear(credential)
	err = handler.application.CompleteWebAuthn(r.Context(), claims,
		LifecycleWebAuthnCompleteInput{IdempotencyKey: key,
			ChallengeKeyID: document.ChallengeKeyID, AuthenticatorID: authenticatorID,
			ChallengeID: challengeID, ExpectedAuthenticatorRevision: revision,
			Challenge: challenge, Credential: credential})
	if err != nil {
		handler.writeFailure(w, r, err)
		return
	}
	handler.noContent(w)
}
