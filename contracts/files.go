package contracts

import "embed"

// Files is the HTTP contract tree the API reference serves.
//
//go:embed http
//go:embed runtime/v1/*.json
var Files embed.FS
