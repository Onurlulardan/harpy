package executor

import "github.com/google/uuid"

type relationEdge struct {
	source uuid.UUID
	target uuid.UUID
	slot   int16
}

func (edge relationEdge) validate() error {
	if edge.source == uuid.Nil || edge.target == uuid.Nil ||
		edge.source.Variant() != uuid.RFC4122 ||
		edge.target.Variant() != uuid.RFC4122 ||
		edge.slot < 1 || edge.slot > 100 {
		return ErrContract
	}
	return nil
}
