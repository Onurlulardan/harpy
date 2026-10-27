package lifecycledelivery

import (
	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/google/uuid"
)

func artifactID(reference string) (uuid.UUID, error) {
	parsed, err := lifecycledeliveryv1.ParseReference(reference)
	if err != nil {
		return uuid.Nil, err
	}
	return parsed.ArtifactID(), nil
}
