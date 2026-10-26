package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type Hook struct {
	Name  string
	Begin func() error
	Wait  func(context.Context) error
	Force func() error
}

type Hooks struct {
	mu      sync.Mutex
	entries []Hook
	names   map[string]struct{}
	sealed  bool
}

func NewHooks() *Hooks {
	return &Hooks{names: make(map[string]struct{})}
}

func (hooks *Hooks) Register(hook Hook) error {
	if err := validateHook(hook); err != nil {
		return err
	}
	hooks.mu.Lock()
	defer hooks.mu.Unlock()
	if hooks.sealed {
		return fmt.Errorf("lifecycle hooks are sealed")
	}
	if _, exists := hooks.names[hook.Name]; exists {
		return fmt.Errorf("lifecycle hook %q is already registered", hook.Name)
	}
	hooks.names[hook.Name] = struct{}{}
	hooks.entries = append(hooks.entries, hook)
	return nil
}

func (hooks *Hooks) Begin() error {
	entries := hooks.seal()
	var joined error
	for _, hook := range entries {
		if err := hook.Begin(); err != nil {
			joined = errors.Join(joined, fmt.Errorf("begin %s: %w", hook.Name, err))
		}
	}
	return joined
}

func (hooks *Hooks) Wait(ctx context.Context) error {
	entries := hooks.seal()
	errorsByHook := make(chan error, len(entries))
	var group sync.WaitGroup
	for _, hook := range entries {
		group.Add(1)
		go func(current Hook) {
			defer group.Done()
			if err := current.Wait(ctx); err != nil {
				errorsByHook <- fmt.Errorf("wait %s: %w", current.Name, err)
			}
		}(hook)
	}
	group.Wait()
	close(errorsByHook)
	var joined error
	for err := range errorsByHook {
		joined = errors.Join(joined, err)
	}
	return joined
}

func (hooks *Hooks) Force() error {
	entries := hooks.seal()
	var joined error
	for index := len(entries) - 1; index >= 0; index-- {
		hook := entries[index]
		if err := hook.Force(); err != nil {
			joined = errors.Join(joined, fmt.Errorf("force %s: %w", hook.Name, err))
		}
	}
	return joined
}

func (hooks *Hooks) seal() []Hook {
	hooks.mu.Lock()
	defer hooks.mu.Unlock()
	hooks.sealed = true
	return append([]Hook(nil), hooks.entries...)
}

func validateHook(hook Hook) error {
	if hook.Name == "" {
		return fmt.Errorf("lifecycle hook name is required")
	}
	if hook.Begin == nil || hook.Wait == nil || hook.Force == nil {
		return fmt.Errorf("lifecycle hook %q is incomplete", hook.Name)
	}
	return nil
}
