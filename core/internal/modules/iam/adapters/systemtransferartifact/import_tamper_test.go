package systemtransferartifact

import (
	"context"
	"testing"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestImportSourceOpenFailsClosedOnTamperedEvidence(t *testing.T) {
	tests := []struct {
		name   string
		change func(*systemtransfer.SourceArtifact, *fixtureObject)
	}{
		{name: "plaintext digest", change: func(artifact *systemtransfer.SourceArtifact,
			_ *fixtureObject) {
			artifact.PlaintextSHA256[0] ^= 1
		}},
		{name: "header digest", change: func(artifact *systemtransfer.SourceArtifact,
			_ *fixtureObject) {
			artifact.HeaderSHA256[0] ^= 1
		}},
		{name: "ciphertext body", change: func(artifact *systemtransfer.SourceArtifact,
			store *fixtureObject) {
			identity := artifactobject.Identity{Key: artifact.ObjectKey,
				Version: artifact.ObjectVersion}
			store.objects[identity][len(store.objects[identity])-1] ^= 1
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			runtime, imports, _ := newRuntimeFixture(t)
			tenantID, actorID, sourceID := fixtureV7(t), fixtureV7(t), fixtureV7(t)
			var artifact systemtransfer.SourceArtifact
			_, err := runtime.SealImportSource(context.Background(), importRequest(
				tenantID, actorID, sourceID, []byte("email\na@b.test\n")),
				func(_ context.Context, value systemtransfer.SourceArtifact) (
					systemtransfer.SourceRecord, error) {
					artifact = value
					return sourceRecord(sourceID, value, false), nil
				})
			if err != nil {
				t.Fatal(err)
			}
			test.change(&artifact, imports)
			reader, openErr := runtime.OpenExactImportSource(context.Background(),
				systemtransfer.ExactImportSource{TenantID: tenantID, ActorID: actorID,
					SourceID: sourceID, Resource: publicresource.Users,
					Format: "csv", Artifact: artifact})
			if openErr == nil || reader != nil {
				if reader != nil {
					_ = reader.Close()
				}
				t.Fatalf("tampered source opened: %v", openErr)
			}
		})
	}
}
