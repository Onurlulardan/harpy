package profilebackfill

import (
	"errors"
	"testing"
)

func TestCreateRejectsMissingDependencies(t *testing.T) {
	if _, err := NewService(nil, nil); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("NewService() error = %v", err)
	}
}
