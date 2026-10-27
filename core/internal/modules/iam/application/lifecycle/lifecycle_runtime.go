package lifecycle

import (
	"fmt"
	"reflect"
)

type Service struct {
	store   Store
	effects Effects
}

func NewService(store Store, effects Effects) (*Service, error) {
	if nilDependency(store) || nilDependency(effects) {
		return nil, ErrUnavailable
	}
	return &Service{store: store, effects: effects}, nil
}

func (service *Service) Valid() bool {
	return service != nil && !nilDependency(service.store) && !nilDependency(service.effects)
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}

func changeFrom(evidence MutationEvidence, key, kind string, result MutationResult) Change {
	return Change{OperationID: evidence.OperationID, RequestSHA256: evidence.RequestSHA256,
		CommandKey: key, AggregateKind: kind, AggregateID: result.ID,
		ActorID: evidence.ActorID, Revision: result.Revision}
}

func credentialChange(evidence CredentialEvidence, key, kind string, result MutationResult) Change {
	return Change{OperationID: evidence.OperationID, RequestSHA256: evidence.RequestSHA256,
		CommandKey: key, AggregateKind: kind, AggregateID: result.ID,
		ActorID: result.SubjectUserID, Revision: result.Revision}
}

func unavailable(operation string, err error) error {
	if err == nil {
		return fmt.Errorf("%s: %w", operation, ErrUnavailable)
	}
	return err
}
