package adoption

import (
	"context"
	"errors"
	"testing"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

func TestAcceptConsumerRejectsEmptyAcknowledgement(t *testing.T) {
	err := (&store{}).AcceptConsumer(
		context.Background(), adoptionapp.ConsumerAck{}, time.Time{},
	)
	if !errors.Is(err, adoptionapp.ErrInvalid) {
		t.Fatal(err)
	}
}
