package physicalid

import (
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

var encoding = base32.StdEncoding.WithPadding(base32.NoPadding)

var allowedPrefixes = map[string]struct{}{
	"cst_": {},
	"fld_": {},
	"frm_": {},
	"idx_": {},
	"rel_": {},
}

// Name derives the single canonical physical identifier for a logical UUID.
func Name(prefix string, logicalID uuid.UUID) (string, error) {
	if _, allowed := allowedPrefixes[prefix]; !allowed ||
		logicalID == uuid.Nil || logicalID.Variant() != uuid.RFC4122 {
		return "", fmt.Errorf("physical identifier input is invalid")
	}
	encoded := strings.ToLower(encoding.EncodeToString(logicalID[:]))
	return prefix + encoded, nil
}
