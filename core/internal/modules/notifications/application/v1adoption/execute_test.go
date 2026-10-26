package v1adoption

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/google/uuid"
)

type adoptionStoreFake struct {
	run     Run
	records []SourceRecord
	targets map[uuid.UUID][32]byte
	issues  int
}

func (store *adoptionStoreFake) BeginAndLock(context.Context, Plan) (Run, error) {
	return store.run, nil
}
func (*adoptionStoreFake) SourceExists(context.Context) (bool, error) { return true, nil }
func (store *adoptionStoreFake) LoadSourcePage(_ context.Context, after *uuid.UUID,
	limit int,
) (SourcePage, error) {
	start := 0
	if after != nil {
		for index := range store.records {
			if store.records[index].ID == *after {
				start = index + 1
				break
			}
		}
	}
	end := min(start+limit, len(store.records))
	return SourcePage{Records: append([]SourceRecord(nil), store.records[start:end]...),
		Final: end == len(store.records)}, nil
}
func (store *adoptionStoreFake) Adopt(_ context.Context, value MappedRecord) ([32]byte, bool, error) {
	if existing, found := store.targets[value.MessageID]; found {
		return existing, false, nil
	}
	store.targets[value.MessageID] = value.CanonicalDigest
	return value.CanonicalDigest, true, nil
}
func (store *adoptionStoreFake) LogIssue(context.Context, uuid.UUID, uuid.UUID, [32]byte, string) error {
	store.issues++
	return nil
}
func (store *adoptionStoreFake) SaveCheckpoint(_ context.Context, _ uuid.UUID, value Checkpoint) error {
	store.run.Checkpoint = value
	return nil
}
func (store *adoptionStoreFake) SealRun(_ context.Context, _ uuid.UUID, value Checkpoint) error {
	if store.issues != 0 {
		return ErrBlocked
	}
	store.run.Checkpoint, store.run.State = value, "verified"
	return nil
}

func TestExecuteResumesWithoutDuplicateAndIgnoresEphemeralWebsocket(t *testing.T) {
	tenant := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	user := uuid.MustParse("10000000-0000-4000-8000-000000000002")
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	record := func(id, channel string, revision int64) SourceRecord {
		created := now.Add(time.Duration(revision) * time.Minute)
		return SourceRecord{ID: uuid.MustParse(id), TenantID: tenant, UserID: &user,
			NotificationType: "workflow_notification", Title: "Title", Body: "Body",
			Channel: channel, CreatedAt: &created, InboxRevision: revision}
	}
	store := &adoptionStoreFake{run: Run{ID: uuid.MustParse("10000000-0000-4000-8000-000000000003"), State: "applying"},
		records: []SourceRecord{
			record("20000000-0000-4000-8000-000000000001", "flow", 1),
			record("20000000-0000-4000-8000-000000000002", "websocket", 2),
		}, targets: map[uuid.UUID][32]byte{}}
	plan := Plan{RunID: store.run.ID, TenantID: tenant, SourceGeneration: "v1-main",
		SourceManifestDigest: sha256.Sum256([]byte("manifest")), PageSize: 1}
	runner, _ := NewRunner(store)
	first, err := runner.Execute(t.Context(), plan)
	if err != nil || first.Complete || first.AdoptedCount != 1 {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	resumed, err := runner.Execute(t.Context(), plan)
	if err != nil || !resumed.Complete || resumed.AdoptedCount != 1 || len(store.targets) != 1 {
		t.Fatalf("resumed=%+v targets=%d err=%v", resumed, len(store.targets), err)
	}
	replay, err := runner.Execute(t.Context(), plan)
	if err != nil || !replay.Complete || replay.AdoptedCount != 1 {
		t.Fatalf("replay=%+v err=%v", replay, err)
	}
}
