package adoption

import (
	"context"
	"errors"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestAcceptConsumerRejectsEmptyAcknowledgement(t *testing.T) {
	err := (&Application{}).AcceptConsumer(
		context.Background(), platformtx.TenantScope{}, ConsumerAck{},
	)
	if !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
}
