package subscription

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (storage *store) expectID(
	ctx context.Context,
	statement string,
	expected uuid.UUID,
	arguments ...any,
) error {
	var actual uuid.UUID
	if err := storage.queryRow(ctx, statement, arguments...).Scan(&actual); err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("persisted identifier %s does not match %s", actual, expected)
	}
	return nil
}

func optionalInstantValue(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}
