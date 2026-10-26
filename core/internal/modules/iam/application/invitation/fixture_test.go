package invitation

import (
	"bytes"
	"encoding/base64"

	deliveryartifactv1 "github.com/Onurlulardan/harpy/contracts/deliveryartifact/v1"
	"github.com/google/uuid"
)

const invitationFixtureKey = "gcp-kms://projects/harpy-prod/locations/europe-west1/" +
	"keyRings/iam/cryptoKeys/delivery/cryptoKeyVersions/7"

func validTokenIssue(invitationID, tenantID uuid.UUID) TokenIssue {
	tokenID, artifactID := uuid.New(), uuid.New()
	aad, _ := deliveryartifactv1.AADSHA256(
		tenantID, artifactID, invitationID, tokenID, invitationFixtureKey,
	)
	envelope := deliveryartifactv1.NewEnvelope(
		artifactID, invitationID, tokenID, invitationFixtureKey,
		[12]byte{1}, [66]byte{2}, [16]byte{3}, aad,
	)
	reference, _ := deliveryartifactv1.NewReference(tenantID, artifactID)
	return TokenIssue{
		ID: tokenID, Selector: [16]byte{1}, HashKeyID: fixtureTokenKeyID(),
		SecretHMAC: [32]byte{2}, BindingSHA256: [32]byte{3},
		DeliveryArtifactRef: reference.String(), DeliveryEnvelope: envelope,
	}
}

func validCreateCommand() CreateCommand {
	invitationID, tenantID := uuid.New(), uuid.New()
	return CreateCommand{
		OperationID: uuid.New(), RequestSHA256: [32]byte{8},
		ID: invitationID, DisplayValue: "invitee@example.com",
		NormalizedValue: "invitee@example.com", NormalizedSHA256: [32]byte{4},
		RoleID:    uuid.New(),
		RoleScope: RoleScope{Kind: RoleScopeTenant, ID: uuid.New()},
		InvitedBy: uuid.New(), ExpectedAuthorizationVersion: 1,
		Token: validTokenIssue(invitationID, tenantID),
	}
}

func validTokenProof() TokenProof {
	return TokenProof{
		ID: uuid.New(), Selector: [16]byte{1}, HashKeyID: fixtureTokenKeyID(),
		SecretHMAC: [32]byte{2}, BindingSHA256: [32]byte{3}, Revision: 1,
	}
}

func fixtureTokenKeyID() string {
	return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{25}, 12))
}
