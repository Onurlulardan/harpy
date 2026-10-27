package v1query

import (
	"encoding/json"
	"fmt"
	"math"
	"testing"
	"testing/quick"
)

func TestPredecessorOrdinalCheckedArithmeticProperty(t *testing.T) {
	property := func(rawPage uint32, rawLimit uint16) bool {
		page := int64(rawPage%1_000_000) + 2
		limit := int(rawLimit%500) + 1
		plan, err := newPagePlan(page, limit)
		if err != nil {
			return false
		}
		ordinal, present, err := plan.PredecessorOrdinal()
		index, indexErr := ordinal.Index()
		return present && err == nil && indexErr == nil &&
			index == (page-1)*int64(limit)-1
	}
	if err := quick.Check(property, &quick.Config{MaxCount: 1_000}); err != nil {
		t.Fatal(err)
	}
	if _, err := newPagePlan(math.MaxInt64, 500); err == nil {
		t.Fatal("overflowing page plan was accepted")
	}
}

func TestPredecessorOrdinalTamperRedactionAndJSONFailClosed(t *testing.T) {
	plan, _ := newPagePlan(3, 7)
	ordinal, present, err := plan.PredecessorOrdinal()
	if err != nil || !present || ordinal.Validate() != nil {
		t.Fatal("valid ordinal was not minted")
	}
	tampered := ordinal
	tampered.index++
	if tampered.Validate() == nil {
		t.Fatal("tampered ordinal was accepted")
	}
	if fmt.Sprint(ordinal) != v1Redaction {
		t.Fatal("ordinal formatting leaked")
	}
	if document, err := json.Marshal(ordinal); err == nil || document != nil {
		t.Fatal("ordinal JSON marshal did not fail closed")
	}
}
