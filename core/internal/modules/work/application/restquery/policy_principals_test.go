package restquery

import (
	"testing"

	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	"github.com/google/uuid"
)

func TestWorkPrincipalKeysRequireCurrentPrincipalEvidence(t *testing.T) {
	id := uuid.New()
	principals := []workprincipalapi.Principal{{
		Kind: "USER", ID: [16]byte(id), Revision: 1,
	}}
	keys, ok := workPrincipalKeys(principals)
	if !ok || len(keys) != 1 || keys[0] != "USER:"+id.String() {
		t.Fatalf("valid principal evidence rejected: keys=%v ok=%t", keys, ok)
	}
	principals[0].Revision = 0
	if _, ok = workPrincipalKeys(principals); ok {
		t.Fatal("zero-revision principal evidence was accepted")
	}
}
