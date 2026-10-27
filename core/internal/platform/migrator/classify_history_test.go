package migrator

import (
	"errors"
	"strings"
	"testing"
)

func TestClassifyHistoryEnforcesPrefixAndExactMetadata(t *testing.T) {
	selected := runnerTestSelection()
	tests := []struct {
		name    string
		history map[int64]historyRow
		wantErr error
	}{
		{
			name: "canonical prefix",
			history: map[int64]historyRow{
				selected[0].Version: runnerTestHistory(selected[0]),
			},
		},
		{
			name: "out of order",
			history: map[int64]historyRow{
				selected[1].Version: runnerTestHistory(selected[1]),
			},
			wantErr: ErrOutOfOrder,
		},
		{
			name: "hash drift",
			history: map[int64]historyRow{
				selected[0].Version: func() historyRow {
					row := runnerTestHistory(selected[0])
					row.selectedSHA256 = selected[1].SHA256
					return row
				}(),
			},
			wantErr: errors.New("metadata"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current, pending, err := classifyHistory(test.history, selected, nil)
			if test.wantErr != nil {
				if err == nil {
					t.Fatal("invalid history was accepted")
				}
				if test.wantErr == ErrOutOfOrder && !errors.Is(err, ErrOutOfOrder) {
					t.Fatalf("error = %v, want out-of-order", err)
				}
				return
			}
			if err != nil || current != selected[0].Version || len(pending) != 1 {
				t.Fatalf("classification = %d, %d, %v", current, len(pending), err)
			}
		})
	}
}

func TestClassifyHistoryAcceptsOnlyExactSupersededIdentity(t *testing.T) {
	selected := runnerTestSelection()
	current := selected[0]
	oldPath := "migrations/superseded/first.sql"
	oldHash := strings.Repeat("c", 64)
	aliases := map[int64]historySupersession{
		current.Version: {
			version: current.Version, owner: current.Owner, scope: current.Scope,
			origin: current.Origin, currentPath: current.Path,
			currentSHA256: current.SHA256,
			superseded:    []historyIdentity{{path: oldPath, sha256: oldHash}},
		},
	}
	row := runnerTestHistory(current)
	row.selectedPath, row.selectedSHA256 = oldPath, oldHash
	history := map[int64]historyRow{current.Version: row}
	got, pending, err := classifyHistory(history, selected, aliases)
	if err != nil || got != current.Version || len(pending) != 1 {
		t.Fatalf("superseded history classification = %d, %d, %v", got, len(pending), err)
	}
	for _, mutate := range []func(*historyRow){
		func(value *historyRow) { value.selectedPath += ".tampered" },
		func(value *historyRow) { value.selectedSHA256 = strings.Repeat("d", 64) },
	} {
		tampered := row
		mutate(&tampered)
		if _, _, err := classifyHistory(
			map[int64]historyRow{current.Version: tampered}, selected, aliases,
		); err == nil {
			t.Fatal("tampered superseded history was accepted")
		}
	}
}

func TestClassifyHistoryAcceptsEverySupersessionChainIdentity(t *testing.T) {
	selected := runnerTestSelection()
	current := selected[0]
	aliases := map[int64]historySupersession{
		current.Version: {
			version: current.Version, owner: current.Owner, scope: current.Scope,
			origin: current.Origin, currentPath: current.Path,
			currentSHA256: current.SHA256,
			superseded: []historyIdentity{
				{path: "migrations/original.sql", sha256: strings.Repeat("c", 64)},
				{path: "migrations/first-replacement.sql", sha256: strings.Repeat("d", 64)},
			},
		},
	}
	for _, identity := range aliases[current.Version].superseded {
		row := runnerTestHistory(current)
		row.selectedPath, row.selectedSHA256 = identity.path, identity.sha256
		if _, _, err := classifyHistory(
			map[int64]historyRow{current.Version: row}, selected, aliases,
		); err != nil {
			t.Fatalf("chain identity %s was rejected: %v", identity.path, err)
		}
	}
}
