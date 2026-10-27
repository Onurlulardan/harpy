package v1query

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sync/atomic"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

const (
	WarningHeaderName      = "Warning"
	DeprecatedWarningValue = `299 Harpy "Deprecated query syntax"`
	acceptedSealDomain     = "harpy/rest-query/v1-accepted/v1\x00"
)

type Accepted struct {
	plan       resolved.AuthorizedPlan
	page       PagePlan
	deprecated bool
	state      *acceptedState
	stateSeal  [sha256.Size]byte
	validated  bool
}

type acceptedState struct{ consumed atomic.Bool }

func newAccepted(
	plan resolved.AuthorizedPlan, page PagePlan, deprecated bool,
) (Accepted, error) {
	if validateAcceptedPlan(plan, page) != nil {
		return Accepted{}, fmt.Errorf("V1 accepted execution is invalid")
	}
	accepted := Accepted{
		plan: plan, page: page, deprecated: deprecated,
		state: &acceptedState{}, validated: true,
	}
	accepted.stateSeal = sealAccepted(accepted)
	if err := accepted.Validate(); err != nil {
		return Accepted{}, err
	}
	return accepted, nil
}

func (accepted Accepted) Validate() error {
	if !accepted.validated || accepted.state == nil ||
		validateAcceptedPlan(accepted.plan, accepted.page) != nil ||
		accepted.stateSeal == [sha256.Size]byte{} ||
		accepted.stateSeal != sealAccepted(accepted) {
		return fmt.Errorf("V1 accepted execution is invalid")
	}
	return nil
}

// ConsumeExecution releases the plan exactly once after evidence acceptance.
func (accepted Accepted) ConsumeExecution() (
	resolved.AuthorizedPlan, PagePlan, error,
) {
	if err := accepted.Validate(); err != nil {
		return resolved.AuthorizedPlan{}, PagePlan{}, err
	}
	if !accepted.state.consumed.CompareAndSwap(false, true) {
		return resolved.AuthorizedPlan{}, PagePlan{},
			fmt.Errorf("V1 accepted execution was already consumed")
	}
	return accepted.plan, accepted.page, nil
}

func (accepted Accepted) Warning() (string, bool, error) {
	if err := accepted.Validate(); err != nil {
		return "", false, err
	}
	if !accepted.deprecated {
		return "", false, nil
	}
	return DeprecatedWarningValue, true, nil
}

func validateAcceptedPlan(plan resolved.AuthorizedPlan, pagePlan PagePlan) error {
	if plan.Validate() != nil || pagePlan.Validate() != nil {
		return fmt.Errorf("V1 authorized plan is invalid")
	}
	spec, err := plan.Spec()
	if err != nil {
		return err
	}
	record, present, err := spec.Record()
	if err != nil || !present {
		return fmt.Errorf("V1 authorized plan is invalid")
	}
	page, err := record.Page()
	if err != nil {
		return err
	}
	size, sizeErr := page.Size()
	_, hasAfter, afterErr := page.After()
	count, countErr := record.Count()
	includes, includeErr := record.Includes()
	if sizeErr != nil || afterErr != nil || countErr != nil || includeErr != nil ||
		size != pagePlan.limit || hasAfter || count != api.CountModeExact ||
		len(includes) != 0 {
		return fmt.Errorf("V1 authorized plan is invalid")
	}
	return nil
}

func sealAccepted(accepted Accepted) [sha256.Size]byte {
	hash := sha256.New()
	_, _ = hash.Write([]byte(acceptedSealDomain))
	identity, _ := accepted.plan.Spec()
	identityHash, _ := identity.IdentityHash()
	authorizationHash, _ := accepted.plan.AuthorizationContextSHA256()
	tenantID, _ := accepted.plan.TenantID()
	actorID, _ := accepted.plan.ActorID()
	mandatory, _ := accepted.plan.MandatoryCanonical()
	_, _ = hash.Write(identityHash[:])
	_, _ = hash.Write(authorizationHash[:])
	_, _ = hash.Write(tenantID[:])
	_, _ = hash.Write(actorID[:])
	_, _ = hash.Write(mandatory)
	number := make([]byte, 8)
	binary.BigEndian.PutUint64(number, uint64(accepted.page.page))
	_, _ = hash.Write(number)
	binary.BigEndian.PutUint64(number, uint64(accepted.page.limit))
	_, _ = hash.Write(number)
	if accepted.deprecated {
		_, _ = hash.Write([]byte{1})
	} else {
		_, _ = hash.Write([]byte{0})
	}
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (Accepted) String() string                          { return v1Redaction }
func (Accepted) GoString() string                        { return v1Redaction }
func (accepted Accepted) Format(state fmt.State, _ rune) { writeV1Redaction(state) }
func (Accepted) MarshalJSON() ([]byte, error)            { return rejectV1JSON() }
