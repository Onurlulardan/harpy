package smtp

import (
	"errors"
	"testing"

	gomail "github.com/wneessen/go-mail"
)

func TestClassifyDataCloseWithoutStatusAsUnknown(t *testing.T) {
	kind, code, status := classifySendError(&gomail.SendError{Reason: gomail.ErrSMTPDataClose})
	if kind != ResultUnknown || code != "SMTP_OUTCOME_UNKNOWN" || status != 0 {
		t.Fatalf("classification = %d/%s/%d", kind, code, status)
	}
}

func TestClassifyRecipientRejectionAsPermanent(t *testing.T) {
	kind, code, _ := classifySendError(&gomail.SendError{Reason: gomail.ErrSMTPRcptTo})
	if kind != ResultPermanent || code != "SMTP_RECIPIENT_REJECTED" {
		t.Fatalf("classification = %d/%s", kind, code)
	}
}

func TestClassifyConnectionOpenFailureAsTransient(t *testing.T) {
	kind, code, _ := classifySendError(errors.New("dial refused"))
	if kind != ResultTransient || code != "SMTP_UNAVAILABLE" {
		t.Fatalf("classification = %d/%s", kind, code)
	}
}
