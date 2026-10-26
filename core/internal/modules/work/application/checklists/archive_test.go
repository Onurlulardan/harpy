package checklists

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestArchiveChecklistRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.ArchiveChecklistItem(context.Background(), nil, ports.Mutation{}, [16]byte{}, [16]byte{}, 0)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
