package systemtransferartifact

import (
	"errors"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
)

func registrationDefinitelyRejected(err error) bool {
	return errors.Is(err, systemtransfer.ErrControlInvalid) ||
		errors.Is(err, systemtransfer.ErrControlNotFound) ||
		errors.Is(err, systemtransfer.ErrControlConflict) ||
		errors.Is(err, systemtransfer.ErrControlCapacity) ||
		errors.Is(err, systemtransfer.ErrControlExpired)
}

func sourceRecordMatches(
	record systemtransfer.SourceRecord,
	request systemtransfer.SourceSealRequest,
	draft sealedDraft,
) bool {
	if record.SourceID.Version() != 7 || record.Resource != string(request.Resource) ||
		record.Format != request.Format || record.PlaintextBytes != draft.plaintextBytes ||
		record.PlaintextSHA256 != draft.plaintextSHA256 || !validEvidenceTime(record.UploadedAt) {
		return false
	}
	if record.SourceID == request.SourceID {
		return !record.Replay
	}
	return record.Replay
}
