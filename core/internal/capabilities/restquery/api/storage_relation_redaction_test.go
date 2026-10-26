package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestStorageAndRelationFormattingIsRedacted(t *testing.T) {
	table, _ := sqlidentifier.MintDynamic("frm_private_records")
	column, _ := sqlidentifier.MintDynamic("fld_private_value_g1")
	resource, _ := NewResourceStorage(table)
	field, _ := NewFieldStorage(column, nil, nil)
	relation, _ := NewRelationDefinition(
		validRelationID, "private_relation", RelationCardinalityToOne,
		mustRelationStorage(t, validRelationID), mustRelationTarget(t), true, 1,
	)
	relationStorage := mustRelationStorage(t, validRelationID)
	for _, value := range []any{resource, field, relationStorage, relation} {
		stringer := value.(fmt.Stringer)
		goStringer := value.(fmt.GoStringer)
		for _, formatted := range []string{
			stringer.String(), goStringer.GoString(), fmt.Sprint(value),
			fmt.Sprintf("%s", value), fmt.Sprintf("%q", value),
			fmt.Sprintf("%#v", value), fmt.Sprintf("%x", value),
		} {
			if formatted != queryValueRedaction {
				t.Fatalf("%T formatting leaked data: %q", value, formatted)
			}
		}
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		} else if strings.Contains(err.Error(), "private") ||
			strings.Contains(err.Error(), validRelationID) {
			t.Fatalf("%T marshal error leaked data: %v", value, err)
		}
	}
}
