package execution

import "regexp"

var ownerNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
