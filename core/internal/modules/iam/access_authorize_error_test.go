package iam

import (
	"errors"
	"fmt"
	"testing"
)

func TestMapAccessAuthorizeErrorPreservesUnavailableCause(t *testing.T) {
	unavailable := errors.New("module unavailable")
	forbidden := errors.New("module forbidden")
	cause := errors.New("evidence clock missing")
	err := mapAccessAuthorizeError(forbidden, unavailable,
		fmt.Errorf("%w: %w", unavailable, cause))
	if !errors.Is(err, unavailable) || !errors.Is(err, cause) {
		t.Fatalf("wrapped unavailable lost cause: %v", err)
	}
	if err := mapAccessAuthorizeError(forbidden, unavailable, forbidden); err != forbidden {
		t.Fatalf("forbidden mapped to %v", err)
	}
}
