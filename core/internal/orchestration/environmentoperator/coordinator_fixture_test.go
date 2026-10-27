package environmentoperator

import (
	"bytes"
	"context"
	"testing"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

type operatorFixture struct {
	coordinator *Coordinator
	onboarding  *onboardingFixture
	migration   *migrationFixture
	authority   *authorityFixture
	finalizer   *finalizerFixture
	sleepErr    *error
}

func newOperatorFixture(t *testing.T,
	manifest installationv1.Manifest,
) operatorFixture {
	t.Helper()
	createdAt := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	onboarding := &onboardingFixture{}
	migration := &migrationFixture{onboarding: onboarding, ready: true,
		states: []string{"succeeded"}, createdAt: createdAt,
		databaseStep: time.Second}
	authority := &authorityFixture{}
	finalizer := &finalizerFixture{onboarding: onboarding}
	var sleepErr error
	coordinator, err := New(Dependencies{Policy: policyFixture{createdAt},
		Onboarding: onboarding, Migration: migration, Offering: offeringFixture{},
		Authority: authority, Finalizer: finalizer,
		Entropy:    bytes.NewReader(bytes.Repeat([]byte{0x51}, 1024)),
		Sleep:      func(context.Context, time.Duration) error { return sleepErr },
		LeaseOwner: "core-operator-initialize", PollInterval: 250 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = manifest
	return operatorFixture{coordinator, onboarding, migration, authority, finalizer,
		&sleepErr}
}
