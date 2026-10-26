package lifecycle

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
)

func TestHooksUseRegistrationOrderAndReverseForceOrder(t *testing.T) {
	hooks := NewHooks()
	var mu sync.Mutex
	var events []string
	record := func(event string) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, event)
	}
	for _, name := range []string{"public", "worker"} {
		current := name
		err := hooks.Register(Hook{
			Name:  current,
			Begin: func() error { record("begin:" + current); return nil },
			Wait:  func(context.Context) error { record("wait:" + current); return nil },
			Force: func() error { record("force:" + current); return nil },
		})
		if err != nil {
			t.Fatalf("register %s: %v", current, err)
		}
	}
	if err := hooks.Begin(); err != nil {
		t.Fatalf("begin: %v", err)
	}
	if err := hooks.Wait(context.Background()); err != nil {
		t.Fatalf("wait: %v", err)
	}
	if err := hooks.Force(); err != nil {
		t.Fatalf("force: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !reflect.DeepEqual(events[:2], []string{"begin:public", "begin:worker"}) {
		t.Fatalf("begin order = %v", events[:2])
	}
	if !reflect.DeepEqual(events[len(events)-2:], []string{"force:worker", "force:public"}) {
		t.Fatalf("force order = %v", events[len(events)-2:])
	}
}

func TestHooksRejectDuplicateIncompleteAndLateRegistration(t *testing.T) {
	hooks := NewHooks()
	if err := hooks.Register(Hook{Name: "bad"}); err == nil {
		t.Fatal("expected incomplete hook rejection")
	}
	complete := Hook{
		Name: "one", Begin: func() error { return nil },
		Wait: func(context.Context) error { return nil }, Force: func() error { return nil },
	}
	if err := hooks.Register(complete); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := hooks.Register(complete); err == nil {
		t.Fatal("expected duplicate rejection")
	}
	if err := hooks.Begin(); err != nil {
		t.Fatalf("begin: %v", err)
	}
	complete.Name = "two"
	if err := hooks.Register(complete); err == nil {
		t.Fatal("expected sealed registry rejection")
	}
}

func TestHooksJoinWaitErrors(t *testing.T) {
	hooks := NewHooks()
	want := errors.New("wait failed")
	if err := hooks.Register(Hook{
		Name: "worker", Begin: func() error { return nil },
		Wait: func(context.Context) error { return want }, Force: func() error { return nil },
	}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if err := hooks.Wait(context.Background()); !errors.Is(err, want) {
		t.Fatalf("wait error = %v", err)
	}
}
