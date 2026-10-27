package eventsv1

import "fmt"

// DecodeDeadLetterEnvelope validates an exact integration envelope without
// requiring the current producer catalog to still register its event version.
func DecodeDeadLetterEnvelope(document []byte) (Envelope, error) {
	var value Envelope
	if err := decodeStrict(document, &value); err != nil {
		return Envelope{}, err
	}
	policy, err := NewProducerPolicy(
		value.EventType,
		value.EventVersion,
		value.ProducerAuthority.Domain,
		value.PayloadSchemaSHA256,
	)
	if err != nil {
		return Envelope{}, fmt.Errorf("build dead-letter producer policy: %w", err)
	}
	catalog, err := NewProducerCatalog([]ProducerPolicy{policy})
	if err != nil {
		return Envelope{}, fmt.Errorf("build dead-letter producer catalog: %w", err)
	}
	value, err = NewEnvelope(value, catalog)
	if err != nil {
		return Envelope{}, fmt.Errorf("validate dead-letter envelope: %w", err)
	}
	return value, nil
}
