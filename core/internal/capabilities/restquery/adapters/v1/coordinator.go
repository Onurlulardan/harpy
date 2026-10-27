package v1query

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/contractretirement"
)

// AuthorizationPort owns catalog resolution and authorization before alias evidence.
type AuthorizationPort interface {
	ResolveAndAuthorizeV1(context.Context, api.QueryRequest) (
		resolved.AuthorizedPlan, error,
	)
}

type Coordinator struct {
	decoder    decoder
	authorizer AuthorizationPort
	evidence   contractretirement.RuntimePort
	admission  *coordinatorAdmission
	validated  bool
}

type coordinatorAdmission struct {
	mutex    sync.Mutex
	admitted atomic.Bool
}

func NewCoordinator(
	source VerifiedManifestSource,
	authorizer AuthorizationPort,
	evidence contractretirement.RuntimePort,
) (Coordinator, error) {
	decoded, err := newDecoder(source)
	if err != nil || nilInterface(authorizer) || nilInterface(evidence) {
		return Coordinator{}, fmt.Errorf("V1 compatibility coordinator is invalid")
	}
	coordinator := Coordinator{
		decoder: decoded, authorizer: authorizer, evidence: evidence,
		admission: &coordinatorAdmission{}, validated: true,
	}
	if err := coordinator.Validate(); err != nil {
		return Coordinator{}, err
	}
	return coordinator, nil
}

func (coordinator Coordinator) Validate() error {
	if !coordinator.validated || coordinator.decoder.Validate() != nil ||
		nilInterface(coordinator.authorizer) || nilInterface(coordinator.evidence) ||
		coordinator.admission == nil {
		return fmt.Errorf("V1 compatibility coordinator is invalid")
	}
	return nil
}

// Accept orders transport decode, catalog authorization, hit evidence, then minting.
func (coordinator Coordinator) Accept(
	ctx context.Context, rawQuery string,
) (Accepted, error) {
	if ctx == nil || coordinator.Validate() != nil {
		return Accepted{}, fmt.Errorf("V1 compatibility coordinator is invalid")
	}
	if !coordinator.admission.admitted.Load() {
		return Accepted{}, fmt.Errorf("V1 compatibility manifest is not admitted")
	}
	pending, err := coordinator.decoder.Decode(rawQuery)
	if err != nil {
		return Accepted{}, err
	}
	request, err := pending.Request()
	if err != nil {
		return Accepted{}, fmt.Errorf("V1 pending request is invalid")
	}
	plan, err := coordinator.authorizer.ResolveAndAuthorizeV1(ctx, request)
	if err != nil {
		return Accepted{}, err
	}
	if validateAcceptedPlan(plan, pending.page) != nil {
		return Accepted{}, fmt.Errorf("V1 authorized plan is invalid")
	}
	deprecated, err := coordinator.recordHits(ctx, pending)
	if err != nil {
		return Accepted{}, err
	}
	return newAccepted(plan, pending.page, deprecated)
}

func nilInterface(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

func (Coordinator) String() string   { return v1Redaction }
func (Coordinator) GoString() string { return v1Redaction }
func (coordinator Coordinator) Format(state fmt.State, _ rune) {
	writeV1Redaction(state)
}
func (Coordinator) MarshalJSON() ([]byte, error) { return rejectV1JSON() }
