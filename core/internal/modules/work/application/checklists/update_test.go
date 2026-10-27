package checklists

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateChecklistRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateChecklistItem(context.Background(), nil, ports.Mutation{}, UpdateChecklistInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
