package passwordblocklist

import (
	"sync"

	blocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

type Blocklist struct {
	mu       sync.RWMutex
	content  []byte
	artifact blocklistv1.Artifact
	open     bool
}
