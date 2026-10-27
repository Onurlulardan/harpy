package migration

import (
	"fmt"
	"path"
)

func flattenDocuments(documents map[string][]byte) (map[string][]byte, error) {
	flattened := make(map[string][]byte, len(documents))
	for sourcePath, content := range documents {
		basename := path.Base(sourcePath)
		if _, duplicate := flattened[basename]; duplicate {
			return nil, fmt.Errorf("migration basename %s is duplicated", basename)
		}
		flattened[basename] = append([]byte(nil), content...)
	}
	return flattened, nil
}
