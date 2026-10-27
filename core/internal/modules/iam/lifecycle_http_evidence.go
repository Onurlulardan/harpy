package iam

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	"github.com/google/uuid"
)

func lifecycleOperationID(
	claims accesstoken.Claims, operation, idempotencyKey string,
) uuid.UUID {
	return permissionRequestOperationID(claims, operation, idempotencyKey)
}

func lifecycleChildID(parent uuid.UUID, purpose string) uuid.UUID {
	return permissionRequestChildID(parent, "lifecycle-"+purpose)
}

func lifecycleRequestSHA(value any) ([32]byte, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, ErrLifecycleHTTPInvalid
	}
	digest := sha256.Sum256(document)
	clear(document)
	return digest, nil
}

func lifecycleEvidence(
	claims accesstoken.Claims, permissionID, operationID uuid.UUID, requestSHA [32]byte,
) lifecycle.MutationEvidence {
	return lifecycle.MutationEvidence{
		OperationID: operationID, RequestSHA256: requestSHA,
		ActorID: claims.Subject(), PermissionID: permissionID,
		ExpectedTokenVersion:         claims.TokenVersion(),
		ExpectedAuthorizationVersion: claims.AuthorizationVersion(),
	}
}

func lifecycleChallengeContext(
	tenantID, userID, sessionID, authenticatorID, challengeID uuid.UUID,
	kind, rpID, origin string,
) [32]byte {
	hash := sha256.New()
	for _, part := range [][]byte{
		[]byte("harpy:iam:lifecycle-enrollment:v1"), tenantID[:], userID[:],
		sessionID[:], authenticatorID[:], challengeID[:], []byte(kind),
		[]byte(rpID), []byte(origin),
	} {
		var length [4]byte
		binary.BigEndian.PutUint32(length[:], uint32(len(part)))
		_, _ = hash.Write(length[:])
		_, _ = hash.Write(part)
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}
