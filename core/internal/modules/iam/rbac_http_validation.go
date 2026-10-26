package iam

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

var rbacCatalogKeyPattern = regexp.MustCompile(
	`^[a-z][a-z0-9]*([-_][a-z0-9]+)*(\.[a-z][a-z0-9]*([-_][a-z0-9]+)*)*$`,
)

func rbacListValid(input RBACList) bool {
	return input.Limit >= 1 && input.Limit <= 100 &&
		(input.State == "active" || input.State == "retired" || input.State == "all")
}

func rbacCatalogCreateValid(input RBACCatalogCreate) bool {
	return rbacCatalogKeyValid(input.Key) && rbacDisplayValid(input.DisplayName) &&
		utf8.RuneCountInString(input.Description) <= 2000 &&
		identityIdempotencyPattern.MatchString(input.IdempotencyKey)
}

func rbacCatalogUpdateValid(input RBACCatalogUpdate) bool {
	return input.ID != uuid.Nil && input.ExpectedRevision > 0 &&
		(input.Key == "" || rbacCatalogKeyValid(input.Key)) &&
		rbacDisplayValid(input.DisplayName) &&
		utf8.RuneCountInString(input.Description) <= 2000 &&
		identityIdempotencyPattern.MatchString(input.IdempotencyKey)
}

func rbacMutationValid(input RBACMutation) bool {
	return input.ID != uuid.Nil && input.ExpectedRevision > 0 &&
		identityIdempotencyPattern.MatchString(input.IdempotencyKey)
}

func rbacPermissionCreateValid(input RBACPermissionCreate) bool {
	return input.ResourceID != uuid.Nil && input.Target.Valid() &&
		input.TargetID != uuid.Nil && rbacActionIDsValid(input.ActionIDs) &&
		identityIdempotencyPattern.MatchString(input.IdempotencyKey)
}

func rbacPermissionUpdateValid(input RBACPermissionUpdate) bool {
	return rbacMutationValid(input.RBACMutation) && input.ResourceID != uuid.Nil &&
		input.Target.Valid() && input.TargetID != uuid.Nil &&
		rbacActionIDsValid(input.ActionIDs)
}

func rbacActionIDsValid(values []uuid.UUID) bool {
	if len(values) < 1 || len(values) > 100 {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(values))
	for _, value := range values {
		if value == uuid.Nil {
			return false
		}
		if _, duplicate := seen[value]; duplicate {
			return false
		}
		seen[value] = struct{}{}
	}
	return true
}

func rbacCheckValid(input RBACCheck) bool {
	return rbacCatalogKeyValid(input.ResourceKey) &&
		rbacCatalogKeyValid(input.ActionKey) &&
		input.ResourceKey != "ALL" && input.ActionKey != "ALL"
}

func rbacBatchCheckValid(input RBACBatchCheck) bool {
	if len(input.Checks) < 1 || len(input.Checks) > 100 {
		return false
	}
	seen := make(map[string]struct{}, len(input.Checks))
	for _, check := range input.Checks {
		key := check.ResourceKey + ":" + check.ActionKey
		if !rbacCheckValid(check) {
			return false
		}
		if _, duplicate := seen[key]; duplicate {
			return false
		}
		seen[key] = struct{}{}
	}
	return true
}

func rbacEffectiveReadValid(input RBACEffectiveRead) bool {
	return input.Target.Valid() && input.TargetID != uuid.Nil &&
		input.Limit >= 1 && input.Limit <= 100
}

func rbacCatalogKeyValid(value string) bool {
	return len(value) <= 100 && rbacCatalogKeyPattern.MatchString(value)
}

func rbacDisplayValid(value string) bool {
	count := utf8.RuneCountInString(value)
	return count >= 1 && count <= 255 && strings.TrimSpace(value) == value
}
