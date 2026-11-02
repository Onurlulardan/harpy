package async

import (
	"strings"
)

func migrationOriginSource(documents map[string][]byte) string {
	var source strings.Builder
	for _, document := range documents {
		source.Write(document)
	}
	return source.String()
}
