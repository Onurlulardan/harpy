package observability

import "testing"

func TestSnapshotRejectsImpossibleMetricValues(t *testing.T) {
	for name, values := range map[string][5]int64{
		"negative outbox age": {-1, 0, 0, 0, 0},
		"invalid alarm":       {0, 2, 0, 0, 0},
		"negative job lag":    {0, 0, -1, 0, 0},
		"negative parked job": {0, 0, 0, -1, 0},
		"negative outbox":     {0, 0, 0, 0, -1},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewSnapshot(
				values[0], values[1], values[2], values[3], values[4],
			); err == nil {
				t.Fatal("invalid async observation snapshot was accepted")
			}
		})
	}
}
