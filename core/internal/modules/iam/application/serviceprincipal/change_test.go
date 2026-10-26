package serviceprincipal

import (
	"context"
	"errors"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type recordingChangeSink struct {
	change rbacdomain.Change
	err    error
	calls  int
}

func (sink *recordingChangeSink) WriteChange(
	_ context.Context,
	_ transaction.TenantScope,
	change rbacdomain.Change,
) error {
	sink.calls++
	sink.change = change
	return sink.err
}

func TestCommandsRejectMissingChangeSink(t *testing.T) {
	capability, err := ParseCapabilityKey("forms.record.update")
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewCapabilityCatalog([]CapabilityDefinition{{
		Key: capability, Owner: "forms", ResourceKind: "record", TenantScope: true,
	}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewCommands(commandPortStub{}, catalog, nil); !errors.Is(err, ErrEvidenceMissing) {
		t.Fatalf("missing sink error = %v", err)
	}
}

func TestFinishRequiresAtomicRedactedChangeSink(t *testing.T) {
	sink := &recordingChangeSink{}
	commands := &Commands{sink: sink}
	evidence := commandFixtureEvidence()
	result := MutationResult{
		ID: uuid.New(), Revision: 2, AuthorizationClock: 7,
	}
	got, err := commands.finish(
		context.Background(), transaction.TenantScope{}, evidence, result,
		"iam.service_principal.disable", "iam.service_principal.disabled",
	)
	if err != nil || got != result || sink.calls != 1 ||
		sink.change.AggregateID != rbacdomain.ID(result.ID) ||
		sink.change.RequestSHA256 != evidence.RequestSHA256 {
		t.Fatalf("finish = %#v / %#v / %d / %v", got, sink.change, sink.calls, err)
	}
	sink.err = errors.New("audit unavailable")
	if _, err := commands.finish(
		context.Background(), transaction.TenantScope{}, evidence, result,
		"iam.service_principal.disable", "iam.service_principal.disabled",
	); !errors.Is(err, ErrChangeSink) {
		t.Fatalf("sink failure error = %v", err)
	}
	result.Replayed = true
	calls := sink.calls
	if _, err := commands.finish(
		context.Background(), transaction.TenantScope{}, evidence, result,
		"iam.service_principal.disable", "iam.service_principal.disabled",
	); err != nil || sink.calls != calls {
		t.Fatalf("replay called sink: calls=%d error=%v", sink.calls, err)
	}
}
