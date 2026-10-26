package rbac

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	dottedKey = regexp.MustCompile(`^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$`)
	ownerKey  = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
)

func (input ResourceInput) Validate() error {
	if !input.ID.Valid() || !validDotted(input.Key) || !validOwner(input.OwnerDomain) ||
		!validOwner(input.Kind) || !validDisplay(input.DisplayName) ||
		utf8.RuneCountInString(input.Description) > 2000 || input.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (input ActionInput) Validate() error {
	if !input.ID.Valid() || !validDotted(input.Key) || !validDisplay(input.DisplayName) ||
		utf8.RuneCountInString(input.Description) > 2000 || input.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (input PermissionInput) Validate() error {
	shape := input.MatchKind == "exact" && input.ResourceID.Valid() && input.ActionID.Valid() ||
		input.MatchKind == "resource_all_actions" && input.ResourceID.Valid() && !input.ActionID.Valid() ||
		input.MatchKind == "tenant_all" && !input.ResourceID.Valid() && !input.ActionID.Valid()
	if !input.ID.Valid() || !shape || (input.MatchKind != "exact" && !input.Protected) ||
		input.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (command UpdateCatalog) Validate() error {
	if !command.ID.Valid() || !validDisplay(command.DisplayName) ||
		utf8.RuneCountInString(command.Description) > 2000 ||
		command.ExpectedRevision < 1 || command.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (input ImplicationInput) Validate() error {
	if !input.ID.Valid() || !input.SourcePermissionID.Valid() ||
		!input.ImpliedPermissionID.Valid() || input.SourcePermissionID == input.ImpliedPermissionID ||
		!validDotted(input.RetirementContractKey) || input.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func (command RecordHitCommand) Validate() error {
	if !command.ID.Valid() || command.ExpectedRevision < 1 ||
		command.Evidence.Validate() != nil {
		return ErrInvalidMutation
	}
	return nil
}

func validDotted(value string) bool {
	return len(value) <= 100 && dottedKey.MatchString(value)
}

func validOwner(value string) bool {
	return len(value) <= 100 && ownerKey.MatchString(value)
}

func validDisplay(value string) bool {
	count := utf8.RuneCountInString(value)
	return count >= 1 && count <= 255 && strings.TrimSpace(value) == value
}
