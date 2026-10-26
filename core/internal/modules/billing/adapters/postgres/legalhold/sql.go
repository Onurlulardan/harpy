package legalhold

import _ "embed"

//go:embed sql/platform/resolve_authority.sql
var resolveAuthoritySQL string

//go:embed sql/platform/place.sql
var placeSQL string

//go:embed sql/platform/release.sql
var releaseSQL string
