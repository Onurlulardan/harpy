package invoice

import (
	"time"

	"github.com/google/uuid"
)

func validV7(value uuid.UUID) bool {
	return value != uuid.Nil && value.Version() == uuid.Version(7)
}

func validID(value uuid.UUID) bool { return value != uuid.Nil }

func validInstant(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC && value.Nanosecond()%1_000 == 0
}

func validDigest(value [32]byte) bool { return value != [32]byte{} }

func (snapshot DraftSnapshot) validFor(request FinalizeRequest) bool {
	return snapshot.InvoiceID == request.InvoiceID && validV7(snapshot.InvoiceID) &&
		snapshot.Revision > 0 && len(snapshot.Lines) > 0 && validDigest(snapshot.PolicySHA256)
}
