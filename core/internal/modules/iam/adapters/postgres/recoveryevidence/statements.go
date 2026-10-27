package recoveryevidence

import _ "embed"

//go:embed sql/platform/read_owner_anchors.sql
var readOwnerAnchorsSQL string
