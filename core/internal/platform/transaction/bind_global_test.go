package transaction

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func TestBindGlobalRequiresLiveGlobalScope(t *testing.T) {
	binder, err := postgres.NewGlobalOwnerBinder(
		"billing", func(postgres.BoundExecutor) struct{} { return struct{}{} },
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := BindGlobal(GlobalScope{}, binder); !errors.Is(err, ErrScopeClosed) {
		t.Fatalf("closed global bind error = %v", err)
	}
}
