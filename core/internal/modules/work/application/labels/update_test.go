package labels

import (
	"context"
	"errors"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"testing"
)

func TestUpdateLabelRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateLabel(context.Background(), nil, ports.Mutation{}, domain.Label{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
