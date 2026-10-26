package iam

import (
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	"github.com/google/uuid"
)

func identitySystemResourceTerminalDigest(
	resource publicresource.Resource,
	id uuid.UUID,
	expectedRevision int64,
	state string,
) ([32]byte, error) {
	return identityRequestSHA(struct {
		Resource         string
		ID               uuid.UUID
		ExpectedRevision int64
		State            string
	}{string(resource), id, expectedRevision, state})
}
