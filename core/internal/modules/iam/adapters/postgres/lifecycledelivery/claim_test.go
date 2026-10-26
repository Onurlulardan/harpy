package lifecycledelivery

import (
	"context"
	"errors"
	"strings"
	"testing"

	lifecycledeliveryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycledelivery"
)

func TestClaimRejectsNilStore(t *testing.T) {
	var value *store
	_, err := value.Claim(context.Background(), lifecycledeliveryapp.ClaimCommand{})
	if !errors.Is(err, lifecycledeliveryapp.ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestClaimReadsImmutableArtifactWithoutLockPrivilege(t *testing.T) {
	if strings.Contains(readArtifactSQL, "FOR UPDATE") ||
		strings.Contains(readArtifactSQL, "FOR SHARE") {
		t.Fatal("lifecycle delivery artifact read requires a row-lock privilege")
	}
}
