package change

import identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"

type Change = identitydomain.Change
type RecordedChange = identitydomain.RecordedChange
type MutationEvidence = identitydomain.MutationEvidence

var (
	ErrSink   = identitydomain.ErrChangeSink
	NewChange = identitydomain.NewChange
)
