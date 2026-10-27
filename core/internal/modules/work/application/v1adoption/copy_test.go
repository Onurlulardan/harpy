package v1adoption

import (
	"context"
	"testing"
	"time"
)

func TestCopyPreservesStableBoardIdentityAndDescription(t *testing.T) {
	now := time.Date(2026, 7, 1, 10, 0, 0, 0, time.UTC)
	boardID := fixtureID(4)
	board := LegacyBoard{ID: boardID, OrganizationID: fixtureID(5),
		CreatedBy: fixtureID(6), Name: "Delivery", Description: "V1 description",
		CreatedAt: now, UpdatedAt: now}
	target := &fixtureTarget{}
	receipt, err := fixtureService([]SourceRecord{
		fixtureRecord(KindBoard, boardID, board),
	}, target).Copy(context.Background(), fixtureTenantScope(), fixtureRequest())
	if err != nil || receipt.State != "RECONCILING" {
		t.Fatal(err, receipt.State)
	}
	found := false
	for _, record := range target.plan.Records {
		found = found || record.Kind == TargetBoard && record.ID == boardID
	}
	if !found {
		t.Fatal("stable board target was not planned")
	}
}

func TestCopyPersistsBlockedInvalidEnumPlan(t *testing.T) {
	now := time.Now().UTC()
	boardID, columnID := fixtureID(4), fixtureID(7)
	records := []SourceRecord{
		fixtureRecord(KindBoard, boardID, LegacyBoard{ID: boardID,
			OrganizationID: fixtureID(5), CreatedBy: fixtureID(6), Name: "B",
			CreatedAt: now, UpdatedAt: now}),
		fixtureRecord(KindColumn, columnID, LegacyColumn{ID: columnID, BoardID: boardID,
			Name: "Unknown", LinkedStatus: "MAYBE", CreatedAt: now, UpdatedAt: now}),
	}
	target := &fixtureTarget{}
	receipt, err := fixtureService(records, target).Copy(
		context.Background(), fixtureTenantScope(), fixtureRequest())
	if err != ErrBlocked || receipt.State != "BLOCKED" || len(target.plan.Issues) != 1 {
		t.Fatalf("receipt=%+v issues=%d err=%v", receipt, len(target.plan.Issues), err)
	}
}
