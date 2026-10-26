package systemtransfercleanup

import (
	"context"
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func TestFinalizePersistsExactAbsenceEvidence(t *testing.T) {
	target, receipt := cleanupTarget(), cleanupReceipt()
	database := fixtureDatabase{query: func(statement string, arguments ...any) row {
		if statement != completeSQL || len(arguments) != 13 ||
			arguments[0] != "tenant_cleanup" || arguments[1] != target.TenantID ||
			arguments[2] != target.SourceID || arguments[3] != target.ActorID ||
			arguments[4] != string(target.Resource) || arguments[5] != target.ObjectKey ||
			arguments[6] != target.ObjectVersion || arguments[7] != target.LeaseOwnerID ||
			arguments[8] != target.LeaseGeneration ||
			arguments[9] != receipt.ProviderRequestID ||
			arguments[12] != receipt.VerifiedAbsentAt {
			t.Fatalf("complete arguments = %#v", arguments)
		}
		if string(arguments[10].([]byte)) != string(receipt.DeleteResponseSHA256[:]) ||
			string(arguments[11].([]byte)) != string(receipt.AbsenceProbeSHA256[:]) {
			t.Fatal("completion evidence differs")
		}
		return fixtureRow(func(destinations ...any) error {
			*destinations[0].(*bool) = false
			return nil
		})
	}}
	replay, err := boundStore(t, database).PersistSourceCleanupCompletion(
		context.Background(), app.SourceCleanupCompletion{
			Target: target, Receipt: receipt,
		},
	)
	if err != nil || replay {
		t.Fatalf("PersistSourceCleanupCompletion() = %v, %v", replay, err)
	}
}
