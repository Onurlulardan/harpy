package maintenance

import (
	"crypto/sha256"

	"github.com/google/uuid"
)

type directoryActualRow struct {
	keyID          string
	hmac           []byte
	userID         uuid.UUID
	sourceRevision int64
	kind           string
	state          string
	sourceSHA      []byte
}

type directoryChecksumSubject struct {
	id      uuid.UUID
	source  *directorySourceRow
	actuals []directoryActualRow
}

func (subject *directoryChecksumSubject) wipe() {
	if subject.source != nil {
		subject.source.wipe()
	}
	for index := range subject.actuals {
		clear(subject.actuals[index].hmac)
		clear(subject.actuals[index].sourceSHA)
	}
}

func (actual directoryActualRow) valid() bool {
	return validStoredDirectoryKeyID(actual.keyID) && len(actual.hmac) == sha256.Size &&
		actual.userID != uuid.Nil && actual.sourceRevision > 0 &&
		(actual.kind == "email_v1" || actual.kind == "portal_v1") &&
		(actual.state == "active" || actual.state == "tombstoned") &&
		len(actual.sourceSHA) == sha256.Size
}

func validStoredDirectoryKeyID(value string) bool {
	if validDirectoryKeyID(value) {
		return true
	}
	if len(value) < 1 || len(value) > 128 ||
		!((value[0] >= 'a' && value[0] <= 'z') ||
			(value[0] >= '0' && value[0] <= '9')) {
		return false
	}
	for _, current := range value[1:] {
		if current >= 'a' && current <= 'z' || current >= '0' && current <= '9' ||
			current == '.' || current == '_' || current == '-' {
			continue
		}
		return false
	}
	return true
}
