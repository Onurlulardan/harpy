package s3object

import artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"

var (
	ErrConfiguration      = artifactobject.ErrConfiguration
	ErrWorkloadIdentity   = artifactobject.ErrWorkloadIdentity
	ErrVersioningRequired = artifactobject.ErrVersioningRequired
	ErrPrivateACLRequired = artifactobject.ErrPrivateACLRequired
	ErrObjectInput        = artifactobject.ErrObjectInput
	ErrProviderEvidence   = artifactobject.ErrProviderEvidence
)
