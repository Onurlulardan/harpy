package application

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/domain/recoveryset"
)

func signingFixture(
	t *testing.T, now time.Time,
) (recoverysigningv1.Bundle, ed25519.PrivateKey, ed25519.PrivateKey) {
	t.Helper()
	purposes := []recoverysigningv1.Purpose{recoverysigningv1.PurposeRecoveryManifest,
		recoverysigningv1.PurposeRecoveryControl, recoverysigningv1.PurposeRestoreReceipt}
	materials := make([]recoverysigningv1.PublicMaterial, 0, 3)
	keys := make(map[recoverysigningv1.Purpose]ed25519.PrivateKey, 3)
	for _, purpose := range purposes {
		public, private, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		var encoded [32]byte
		copy(encoded[:], public)
		materials = append(materials, recoverysigningv1.PublicMaterial{
			Purpose: purpose, PublicKey: encoded})
		keys[purpose] = private
	}
	bundle, err := recoverysigningv1.Build(appUUID(1), now.Add(-time.Minute), materials)
	if err != nil {
		t.Fatal(err)
	}
	return bundle, keys[recoverysigningv1.PurposeRecoveryManifest],
		keys[recoverysigningv1.PurposeRecoveryControl]
}

func appIdentity() recoveryset.EnvironmentIdentity {
	return recoveryset.EnvironmentIdentity{EnvironmentID: appUUID(1),
		DeploymentMode: recoveryset.ModeOnPremise, PrimaryBindingID: appUUID(2),
		TenantSetSHA256: appDigest("a"), OwnerBindingID: appUUID(3),
		OwnerBindingGeneration: 4, DatabaseLineageID: appUUID(4)}
}

func appAPIIdentity() api.EnvironmentIdentity {
	identity := appIdentity()
	return api.EnvironmentIdentity{EnvironmentID: identity.EnvironmentID,
		DeploymentMode:         api.DeploymentMode(identity.DeploymentMode),
		PrimaryBindingID:       identity.PrimaryBindingID,
		TenantSetSHA256:        identity.TenantSetSHA256,
		OwnerBindingID:         identity.OwnerBindingID,
		OwnerBindingGeneration: identity.OwnerBindingGeneration,
		DatabaseLineageID:      identity.DatabaseLineageID}
}

func appManifest(identity recoveryset.EnvironmentIdentity, now time.Time) recoveryset.ManifestPayload {
	return recoveryset.ManifestPayload{Schema: recoveryset.ManifestSchema,
		Portability: recoveryset.SameEnvironment, Identity: identity,
		DatabaseSystemID: appUUID(9), DatabaseArchiveSHA256: appDigest("b"),
		DocumentsMerkleSHA256: appDigest("c"), KeyCustodySHA256: appDigest("d"),
		OwnerAnchorSHA256: appDigest("e"), ProtectedHeadAckSHA256: appDigest("f"),
		RecoveryTrustBundleSHA256: appDigest("1"), RecoverySigningGeneration: 1,
		Components: recoveryset.RecoveryComponents{
			DatabaseArchive:    componentFixture("db-archive", appDigest("b")),
			DocumentsInventory: componentFixture("documents-inventory", appDigest("c")),
			KeyCustody:         componentFixture("key-custody", appDigest("d")),
			OwnerAnchor:        componentFixture("owner-anchor", appDigest("e")),
			ProtectedAuthority: componentFixture("protected-authority", appDigest("f")),
		},
		CreatedAt: now}
}

func componentFixture(key, sha string) recoveryset.ComponentIdentity {
	return recoveryset.ComponentIdentity{ObjectKey: key,
		ObjectVersionID: "1", CiphertextSHA256: sha, CiphertextBytes: 1}
}

func appUUID(marker int) string {
	return "018f0000-0000-7000-8000-00000000000" + string(rune('0'+marker))
}

func appDigest(marker string) string {
	value := ""
	for len(value) < 64 {
		value += marker
	}
	return value
}
