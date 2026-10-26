package systemresource

import (
	"encoding/json"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func putValue(
	values map[publicresource.ID]json.RawMessage,
	visible map[publicresource.ID]struct{},
	fieldID string,
	value any,
) error {
	parsed, err := publicresource.ParseID(fieldID)
	if err != nil {
		return publicresource.ErrContract
	}
	if _, allowed := visible[parsed]; !allowed {
		return nil
	}
	if _, duplicate := values[parsed]; duplicate {
		return publicresource.ErrContract
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return publicresource.ErrContract
	}
	values[parsed] = encoded
	return nil
}

func putIdentityValue(
	values map[publicresource.ID]json.RawMessage,
	visible map[publicresource.ID]struct{},
	fieldID string,
	value identityfacade.IdentityID,
) error {
	if value == (identityfacade.IdentityID{}) {
		return putValue(values, visible, fieldID, nil)
	}
	return putValue(values, visible, fieldID, uuid.UUID(value).String())
}

func identityTexts(values []identityfacade.IdentityID) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = uuid.UUID(value).String()
	}
	return result
}

func timestampText(value time.Time) (string, error) {
	if value.IsZero() {
		return "", publicresource.ErrContract
	}
	return value.UTC().Truncate(time.Microsecond).
		Format("2006-01-02T15:04:05.000000Z"), nil
}
