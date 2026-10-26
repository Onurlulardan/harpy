package postgres

import (
	"fmt"

	"github.com/google/uuid"
)

func assignWorkerTestValue(destination any, value any) error {
	switch target := destination.(type) {
	case *string:
		resolved, ok := value.(string)
		if ok {
			*target = resolved
			return nil
		}
	case *uuid.UUID:
		resolved, ok := value.(uuid.UUID)
		if ok {
			*target = resolved
			return nil
		}
	case *int64:
		resolved, ok := value.(int64)
		if ok {
			*target = resolved
			return nil
		}
	case *bool:
		resolved, ok := value.(bool)
		if ok {
			*target = resolved
			return nil
		}
	}
	return fmt.Errorf("worker test value %T cannot scan into %T", value, destination)
}
