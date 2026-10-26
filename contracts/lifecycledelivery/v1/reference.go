package lifecycledeliveryv1

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

const referencePrefix = "iam-lifecycle-delivery://tenant/"

var ErrReference = errors.New("IAM lifecycle delivery reference is invalid")

type Reference struct{ tenantID, artifactID uuid.UUID }

func NewReference(tenantID, artifactID uuid.UUID) (Reference, error) {
	if tenantID == uuid.Nil || artifactID == uuid.Nil {
		return Reference{}, ErrReference
	}
	return Reference{tenantID: tenantID, artifactID: artifactID}, nil
}

func ParseReference(value string) (Reference, error) {
	if len(value) != len(referencePrefix)+36+len("/artifact/")+36 ||
		!strings.HasPrefix(value, referencePrefix) {
		return Reference{}, ErrReference
	}
	parts := strings.Split(strings.TrimPrefix(value, referencePrefix), "/artifact/")
	if len(parts) != 2 {
		return Reference{}, ErrReference
	}
	tenantID, tenantErr := uuid.Parse(parts[0])
	artifactID, artifactErr := uuid.Parse(parts[1])
	if tenantErr != nil || artifactErr != nil || tenantID.String() != parts[0] ||
		artifactID.String() != parts[1] {
		return Reference{}, ErrReference
	}
	return NewReference(tenantID, artifactID)
}

func (reference Reference) String() string {
	if reference.Validate() != nil {
		return ""
	}
	return referencePrefix + reference.tenantID.String() +
		"/artifact/" + reference.artifactID.String()
}

func (reference Reference) Validate() error {
	if reference.tenantID == uuid.Nil || reference.artifactID == uuid.Nil {
		return ErrReference
	}
	return nil
}

func (reference Reference) TenantID() uuid.UUID   { return reference.tenantID }
func (reference Reference) ArtifactID() uuid.UUID { return reference.artifactID }
