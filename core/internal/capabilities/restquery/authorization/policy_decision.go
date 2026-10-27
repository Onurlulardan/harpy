package authorization

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"slices"
)

const maximumPolicyVisibilities = 4

type PolicyDecision struct {
	requestBinding  [sha256.Size]byte
	contextRevision [sha256.Size]byte
	results         []DecisionResult
	visibilities    []Visibility
	validated       bool
}

func NewPolicyDecision(
	requestBinding [sha256.Size]byte,
	contextRevision [sha256.Size]byte,
	results []DecisionResult,
	visibilities []Visibility,
) (PolicyDecision, error) {
	ownedResults := append([]DecisionResult{}, results...)
	slices.SortFunc(ownedResults, compareDecisionResults)
	ownedVisibilities := make([]Visibility, len(visibilities))
	for index := range visibilities {
		ownedVisibilities[index] = visibilities[index].clone()
	}
	slices.SortFunc(ownedVisibilities, compareVisibilities)
	decision := PolicyDecision{
		requestBinding: requestBinding, contextRevision: contextRevision,
		results: ownedResults, visibilities: ownedVisibilities, validated: true,
	}
	if err := decision.Validate(); err != nil {
		return PolicyDecision{}, err
	}
	return decision, nil
}

func (decision PolicyDecision) Validate() error {
	if !decision.validated || decision.requestBinding == [sha256.Size]byte{} ||
		decision.contextRevision == [sha256.Size]byte{} || len(decision.results) < 1 ||
		len(decision.results) > maximumPolicyRequirements ||
		len(decision.visibilities) > maximumPolicyVisibilities ||
		!validDecisionResults(decision.results) ||
		!validDecisionVisibilities(decision.visibilities) {
		return fmt.Errorf("REST Query policy decision is invalid")
	}
	return nil
}

func (PolicyDecision) String() string   { return authorizationRedaction }
func (PolicyDecision) GoString() string { return authorizationRedaction }
func (PolicyDecision) LogValue() slog.Value {
	return authorizationLogValue()
}
func (decision PolicyDecision) Format(state fmt.State, _ rune) {
	writeAuthorizationRedaction(state)
}
func (PolicyDecision) MarshalJSON() ([]byte, error) { return rejectAuthorizationJSON() }

func compareDecisionResults(left, right DecisionResult) int {
	return bytes.Compare(left.requirement[:], right.requirement[:])
}

func compareVisibilities(left, right Visibility) int {
	return bytes.Compare(left.resourceRequirement[:], right.resourceRequirement[:])
}
