package postgres

import "regexp"

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
