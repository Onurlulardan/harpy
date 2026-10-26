//go:build integration

package executor

import "testing"

func (gate *pg18Gate) verifyGraphExecution(t *testing.T) {
	pg18PrepareGraphSchema(t, gate)
	t.Run("canonical relation postconditions and concurrent caps", func(t *testing.T) {
		pg18VerifyGraphStorage(t, gate)
	})
	t.Run("owners snapshot ordering exact count and tenant isolation", func(t *testing.T) {
		pg18VerifyGraphSnapshot(t, gate)
	})
	t.Run("empty frontiers are explicit no-op stages", func(t *testing.T) {
		pg18VerifyGraphEmptyFrontiers(t, gate)
	})
	t.Run("include predicate is bound", func(t *testing.T) {
		pg18VerifyGraphInjection(t, gate)
	})
	t.Run("to-one null and visible cardinality violation", func(t *testing.T) {
		pg18VerifyGraphCardinality(t, gate)
	})
	t.Run("one thousand and first occurrence is atomic 422", func(t *testing.T) {
		pg18VerifyGraphOccurrenceLimit(t, gate)
	})
}
