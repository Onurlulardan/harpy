package projections

import (
	"context"
	"errors"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"testing"
)

func TestRepairProjectionRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.RepairProjection(context.Background(), nil, ports.Mutation{}, domain.SourceOwner(""), "", [16]byte{}, 0, nil)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
