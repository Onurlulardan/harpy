package tenantsettings

import (
	"context"
	"errors"
	"reflect"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errMailInterrupted = errors.New("mail replacement interrupted")

type resumeRegistryFake struct {
	calls     int
	mutations []RegistryMutation
}

func (*resumeRegistryFake) Read(context.Context,
	platformtx.TenantScope,
) (RegistrySnapshot, error) {
	return RegistrySnapshot{}, nil
}

func (fake *resumeRegistryFake) Update(_ context.Context,
	_ platformtx.TenantScope, mutation RegistryMutation,
) (UpdateResult, error) {
	fake.calls++
	fake.mutations = append(fake.mutations, mutation)
	return UpdateResult{Revision: 8, Replay: fake.calls > 1}, nil
}

type replaySafeMailFake struct {
	calls     int
	effects   int
	mutations []MailMutation
	completed *MailMutation
}

func (fake *replaySafeMailFake) Replace(_ context.Context,
	_ platformtx.TenantScope, mutation MailMutation,
) error {
	fake.calls++
	mutation.SMTP.Password = append([]byte(nil), mutation.SMTP.Password...)
	fake.mutations = append(fake.mutations, mutation)
	if fake.calls == 1 {
		return errMailInterrupted
	}
	if fake.completed != nil {
		if !reflect.DeepEqual(*fake.completed, mutation) {
			return ErrConflict
		}
		return nil
	}
	fake.effects++
	fake.completed = &mutation
	return nil
}

func (*replaySafeMailFake) EnqueueTest(context.Context,
	platformtx.TenantScope, TestMailRequest,
) (TestEmailResult, error) {
	return TestEmailResult{}, nil
}

func TestUpdateResumesMailAfterRegistryReplayWithoutDuplicateEffect(t *testing.T) {
	registry, mail := &resumeRegistryFake{}, &replaySafeMailFake{}
	application := &Application{registry: registry, mail: mail}
	command := func() UpdateCommand {
		name := "Acme Updated"
		return UpdateCommand{Name: &name, ExpectedRevision: 7,
			IdempotencyKey: "tenant-settings-0001", SMTP: &SMTPMutation{
				Host: "smtp.example.test", Port: 587, Username: "mailer",
				Password: []byte("smtp-secret"), From: "noreply@example.test"}}
	}

	if _, err := application.applyUpdate(t.Context(), platformScopeZero,
		command()); !errors.Is(err, errMailInterrupted) {
		t.Fatalf("first update error = %v", err)
	}
	second, err := application.applyUpdate(t.Context(), platformScopeZero, command())
	if err != nil || !second.Replay || second.Revision != 8 {
		t.Fatalf("resumed update = %+v, error = %v", second, err)
	}
	third, err := application.applyUpdate(t.Context(), platformScopeZero, command())
	if err != nil || !third.Replay || mail.calls != 3 || mail.effects != 1 {
		t.Fatalf("completed replay = %+v, error = %v, calls = %d, effects = %d",
			third, err, mail.calls, mail.effects)
	}
	if registry.calls != 3 || len(mail.mutations) != 3 {
		t.Fatalf("registry calls = %d, mail mutations = %d",
			registry.calls, len(mail.mutations))
	}
	for index, mutation := range mail.mutations {
		if mutation.ExpectedRevision != 7 || mutation.ResultRevision != 8 ||
			mutation.IdempotencyKey != "tenant-settings-0001" ||
			mutation.RequestSHA256 != registry.mutations[index].RequestSHA256 ||
			!reflect.DeepEqual(mutation, mail.mutations[0]) {
			t.Fatalf("mail mutation %d drifted: %+v", index, mutation)
		}
	}
}
