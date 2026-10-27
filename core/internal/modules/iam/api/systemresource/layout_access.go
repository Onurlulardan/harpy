package systemresource

import (
	"errors"
	"fmt"
)

var (
	ErrLayoutAccessAuthentication = errors.New("IAM system layout authentication failed")
	ErrLayoutAccessForbidden      = errors.New("IAM system layout access is forbidden")
	ErrLayoutAccessUnavailable    = errors.New("IAM system layout access is unavailable")
)

type LayoutAccessRequest struct {
	AuthorizationValues []string
	TenantIDValues      []string
	TenantSlugValues    []string
	Resource            Resource
	Action              string
	Operation           string
}

type LayoutAccessDecision struct {
	admission Admission
}

func NewLayoutAccessDecision(admission Admission) (LayoutAccessDecision, error) {
	if admission.Validate() != nil {
		return LayoutAccessDecision{}, fmt.Errorf("IAM system layout decision is invalid")
	}
	return LayoutAccessDecision{admission: admission}, nil
}

func (decision LayoutAccessDecision) Admission() (Admission, error) {
	if decision.admission.Validate() != nil {
		return Admission{}, ErrLayoutAccessUnavailable
	}
	return decision.admission, nil
}
