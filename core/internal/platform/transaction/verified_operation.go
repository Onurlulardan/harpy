package transaction

import (
	"regexp"
	"slices"
)

const (
	maximumCommandGroups  = 8
	maximumCommandDomains = 16
)

var operationIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,127}$`)
var authorityNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)

type groupExpectation struct {
	id         string
	generation int64
}

type domainExpectation struct {
	name       string
	generation int64
	groupID    string
}

type VerifiedOperation struct {
	operationID  string
	groups       []groupExpectation
	domains      []domainExpectation
	participants []string
	mode         ScopeMode
	issued       bool
}

type AuthorityIssuer struct{ valid bool }

func NewAuthorityIssuer() AuthorityIssuer {
	return AuthorityIssuer{valid: true}
}

func (operation VerifiedOperation) valid() bool {
	if !operation.issued || !operationIDPattern.MatchString(operation.operationID) {
		return false
	}
	if operation.mode == SnapshotQuery {
		return validSnapshotOperation(operation)
	}
	return operation.mode == Command && validCommandOperation(operation)
}

func (operation VerifiedOperation) allowsOwner(owner string) bool {
	if operation.mode == Command {
		return slices.Contains(operation.participants, owner)
	}
	for _, domain := range operation.domains {
		if domain.name == owner {
			return true
		}
	}
	return false
}

func (operation VerifiedOperation) groupForOwner(owner string) (groupExpectation, bool) {
	if !operation.allowsOwner(owner) {
		return groupExpectation{}, false
	}
	for _, domain := range operation.domains {
		if domain.name != owner {
			continue
		}
		for _, group := range operation.groups {
			if group.id == domain.groupID {
				return group, true
			}
		}
	}
	return groupExpectation{}, false
}

func (operation VerifiedOperation) singleGroup() (groupExpectation, bool) {
	if len(operation.groups) != 1 {
		return groupExpectation{}, false
	}
	return operation.groups[0], true
}
