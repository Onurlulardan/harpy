package permissioncachepostgres

import (
	"testing"

	permissioncacheapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissioncache"
)

func TestPurgeRejectsClosedScope(t *testing.T) {
	storage, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = storage.Purge(t.Context(), permissioncacheapp.Command{}); err != permissioncacheapp.ErrInvalid {
		t.Fatalf("Purge() error = %v", err)
	}
}
