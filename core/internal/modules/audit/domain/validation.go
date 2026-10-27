package domain

import (
	"crypto/sha256"
	"regexp"
	"strings"
	"unicode/utf8"
)

var tokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,127}$`)

func (entry Entry) Validate() error {
	if !entry.ID.UUIDv7() || !entry.OwnerAuditID.UUIDv7() || !entry.ActorID.Valid() ||
		!optionalID(entry.OrganizationID) || !optionalID(entry.CorrelationID) ||
		!token(entry.OwnerModule, 64) || !token(entry.Operation, 128) ||
		!token(entry.EntityType, 128) ||
		!text(entry.EntityID, 1, 200) || !optionalToken(entry.FormSlug, 100) ||
		!validAction(entry.Action) || entry.AuthorizationVersion < 1 ||
		!optionalToken(entry.ReasonCode, 128) || len(entry.FieldChanges) > 100 ||
		entry.EventSHA256 == [32]byte{} || !validTime(entry.OccurredAt) ||
		!validTime(entry.RecordedAt) || entry.RecordedAt.Before(entry.OccurredAt) {
		return ErrInvalid
	}
	seen := make(map[string]struct{}, len(entry.FieldChanges))
	for _, change := range entry.FieldChanges {
		if change.Validate() != nil {
			return ErrInvalid
		}
		if _, exists := seen[change.Field]; exists {
			return ErrInvalid
		}
		seen[change.Field] = struct{}{}
	}
	return nil
}

func (change FieldChange) Validate() error {
	if !token(change.Field, 128) || !validClassification(change.Classification) {
		return ErrInvalid
	}
	if change.Classification == ClassificationConfidential ||
		change.Classification == ClassificationSecret {
		if change.Before != nil || change.After != nil ||
			(change.BeforeSHA256 == [32]byte{} && change.AfterSHA256 == [32]byte{}) ||
			change.BeforeSHA256 == change.AfterSHA256 {
			return ErrInvalid
		}
		return nil
	}
	if !displayMatches(change.Before, change.BeforeSHA256) ||
		!displayMatches(change.After, change.AfterSHA256) ||
		change.BeforeSHA256 == change.AfterSHA256 {
		return ErrInvalid
	}
	return nil
}

func (filter Filter) Validate() error {
	if filter.Page < 1 || filter.Page > 1000 || filter.Limit < 1 || filter.Limit > 100 ||
		!optionalToken(filter.OwnerModule, 64) || !optionalToken(filter.EntityType, 128) ||
		!optionalText(filter.EntityID, 200) || !optionalToken(filter.FormSlug, 100) ||
		(filter.Action != "" && !validAction(filter.Action)) ||
		!optionalID(filter.ActorID) || !optionalID(filter.OrganizationID) ||
		!optionalID(filter.CorrelationID) || !optionalText(filter.Search, 128) ||
		(filter.Start != nil && !validTime(*filter.Start)) ||
		(filter.End != nil && !validTime(*filter.End)) ||
		(filter.Start != nil && filter.End != nil && filter.End.Before(*filter.Start)) {
		return ErrInvalid
	}
	return nil
}

func optionalID(value *ID) bool { return value == nil || value.Valid() }

func token(value string, maximum int) bool {
	return len(value) <= maximum && tokenPattern.MatchString(value)
}

func optionalToken(value string, maximum int) bool {
	return value == "" || token(value, maximum)
}

func text(value string, minimum, maximum int) bool {
	return utf8.ValidString(value) && len(value) >= minimum && len([]byte(value)) <= maximum &&
		!strings.ContainsRune(value, 0)
}

func optionalText(value string, maximum int) bool { return value == "" || text(value, 1, maximum) }

func displayMatches(value *string, digest [32]byte) bool {
	if value == nil {
		return digest == [32]byte{}
	}
	return text(*value, 0, 2048) && sha256.Sum256([]byte(*value)) == digest
}

func validClassification(value Classification) bool {
	return value == ClassificationPublic || value == ClassificationInternal ||
		value == ClassificationConfidential || value == ClassificationSecret
}

func validAction(value Action) bool {
	switch value {
	case ActionCreate, ActionUpdate, ActionDelete, ActionTransition, ActionPublish,
		ActionExecute, ActionGrant, ActionRevoke, ActionRestore, ActionPurge:
		return true
	default:
		return false
	}
}

func validTime(value interface {
	IsZero() bool
	Nanosecond() int
}) bool {
	return !value.IsZero() && value.Nanosecond()%1000 == 0
}
