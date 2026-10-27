package iam

import (
	"context"
	"errors"
	"testing"
)

func TestPortalSubjectLinkProjectionFailsClosedWithoutRuntime(t *testing.T) {
	_, err := (*FormsRecordAccessRuntime)(nil).ProjectPortalSubjectLinks(
		context.Background(), PortalSubjectLinkProjectionInput{},
	)
	if !errors.Is(err, ErrFormsRecordAccessUnavailable) {
		t.Fatalf("ProjectPortalSubjectLinks error = %v", err)
	}
}
