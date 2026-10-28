//go:build harpy_measure_no_release_embed

package migration

import "fmt"

func LoadReleaseRegistry() (*ReleaseRegistry, error) {
	return nil, fmt.Errorf("release archive embed excluded by harpy_measure_no_release_embed")
}
