package legalhold

import _ "embed"

//go:embed sql/tenant/place.sql
var placeSQL string

//go:embed sql/tenant/release.sql
var releaseSQL string
