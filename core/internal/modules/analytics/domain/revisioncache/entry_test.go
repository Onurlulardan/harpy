package revisioncache

import (
	"testing"
	"time"
)

func TestCacheHitRejectsGenerationAuthorizationAndRevisionDrift(t *testing.T) {
	binding := fixtureBinding()
	now := time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)
	entry, err := NewEntry(binding, []byte(`{"data":[]}`), now, now.Add(MaximumTTL))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := entry.Hit(binding, now.Add(time.Second)); !ok {
		t.Fatal("current binding missed")
	}
	for name, mutate := range map[string]func(*Binding){
		"generation":    func(value *Binding) { value.ReleaseGeneration++ },
		"authorization": func(value *Binding) { value.AuthorizationSHA256[0]++ },
		"revision":      func(value *Binding) { value.RevisionsSHA256[0]++ },
	} {
		t.Run(name, func(t *testing.T) {
			changed := binding
			mutate(&changed)
			if _, ok := entry.Hit(changed, now.Add(time.Second)); ok {
				t.Fatal("stale binding hit")
			}
		})
	}
	if _, ok := entry.Hit(binding, now.Add(MaximumTTL)); ok {
		t.Fatal("expired binding hit")
	}
}

func fixtureBinding() Binding {
	return Binding{TenantID: [16]byte{1}, ReleaseID: [16]byte{2}, ReleaseGeneration: 3,
		QuerySHA256: [32]byte{1}, ParametersSHA256: [32]byte{2},
		AuthorizationSHA256: [32]byte{3}, PredicateSHA256: [32]byte{4},
		RevisionsSHA256: [32]byte{5}}
}
