package admission

import (
	"testing"
	"time"
)

func TestGateCannotReopenAfterDrain(t *testing.T) {
	gate := &Gate{}
	if gate.Accepting() || gate.Draining() {
		t.Fatal("zero gate must be closed")
	}
	if !gate.Open() || !gate.Accepting() {
		t.Fatal("gate did not open")
	}
	gate.Pause()
	if gate.Accepting() || !gate.Open() {
		t.Fatal("paused gate did not reopen")
	}
	gate.BeginDrain()
	if !gate.Draining() || gate.Accepting() || gate.Open() {
		t.Fatal("draining gate reopened")
	}
}

func TestGateIssuesPermitsOnlyAcrossOpenBoundary(t *testing.T) {
	gate := &Gate{}
	if permit, admitted := gate.Admit(); admitted || permit.Granted() {
		t.Fatal("closed gate issued a permit")
	}
	gate.Open()
	permit, admitted := gate.Admit()
	if !admitted || !permit.Granted() {
		t.Fatal("open gate did not issue a permit")
	}
	gate.BeginDrain()
	if permit, admitted = gate.Admit(); admitted || permit.Granted() {
		t.Fatal("draining gate issued a permit")
	}
}

func TestGateExpiresReadinessAndAllowsRecovery(t *testing.T) {
	now := time.Date(2026, 7, 14, 8, 0, 0, 0, time.UTC)
	gate := NewGate(func() time.Time { return now })
	if !gate.OpenUntil(now.Add(time.Minute)) || !gate.Accepting() {
		t.Fatal("fresh readiness did not open gate")
	}
	now = now.Add(time.Minute + time.Nanosecond)
	if permit, admitted := gate.Admit(); admitted || permit.Granted() || gate.Accepting() {
		t.Fatal("stale readiness still admitted work")
	}
	if !gate.OpenUntil(now.Add(time.Minute)) || !gate.Accepting() {
		t.Fatal("fresh recovery did not reopen gate")
	}
	now = now.Add(-2 * time.Minute)
	if gate.Accepting() {
		t.Fatal("clock rollback kept readiness admission open")
	}
}
