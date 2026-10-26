package providerapp

import (
	"crypto/sha256"
	"regexp"
	"time"
)

const maximumNormalizedFactBytes = 64 << 10

var normalizedSchemaPattern = regexp.MustCompile(`^[a-z][a-z0-9_.]{2,127}$`)

type NormalizedEvent struct {
	schema     string
	revision   int
	occurredAt time.Time
	ordering   EventOrdering
	fact       []byte
	factSHA256 [32]byte
}

func NewNormalizedEvent(
	schema string,
	revision int,
	occurredAt time.Time,
	ordering EventOrdering,
	canonicalFact []byte,
) (NormalizedEvent, error) {
	canonicalOccurred, err := canonicalInstant(occurredAt)
	if err != nil || !normalizedSchemaPattern.MatchString(schema) || revision < 1 ||
		!ordering.valid() || len(canonicalFact) == 0 ||
		len(canonicalFact) > maximumNormalizedFactBytes {
		return NormalizedEvent{}, ErrNormalizedEvent
	}
	return NormalizedEvent{
		schema: schema, revision: revision, occurredAt: canonicalOccurred,
		ordering: ordering, fact: append([]byte(nil), canonicalFact...),
		factSHA256: sha256.Sum256(canonicalFact),
	}, nil
}

func (event NormalizedEvent) Schema() string          { return event.schema }
func (event NormalizedEvent) Revision() int           { return event.revision }
func (event NormalizedEvent) OccurredAt() time.Time   { return event.occurredAt }
func (event NormalizedEvent) Ordering() EventOrdering { return event.ordering }
func (event NormalizedEvent) CanonicalFact() []byte   { return append([]byte(nil), event.fact...) }
func (event NormalizedEvent) FactSHA256() [32]byte    { return event.factSHA256 }

func (event NormalizedEvent) valid() bool {
	validated, err := NewNormalizedEvent(
		event.schema, event.revision, event.occurredAt, event.ordering, event.fact,
	)
	return err == nil && normalizedEventsEqual(validated, event)
}

func normalizedEventsEqual(left, right NormalizedEvent) bool {
	return left.schema == right.schema && left.revision == right.revision &&
		left.occurredAt.Equal(right.occurredAt) && left.ordering == right.ordering &&
		left.factSHA256 == right.factSHA256 && string(left.fact) == string(right.fact)
}
