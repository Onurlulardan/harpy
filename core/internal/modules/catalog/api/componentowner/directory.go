package componentowner

import (
	"fmt"
	"reflect"
)

var requiredOwners = []string{"analytics", "catalog", "iam"}

type StaticDirectory struct{ owners map[string]Owner }

func NewDirectory(owners map[string]Owner) (*StaticDirectory, error) {
	if len(owners) != len(requiredOwners) {
		return nil, fmt.Errorf("Catalog component owner directory is incomplete")
	}
	copyOwners := make(map[string]Owner, len(owners))
	for _, name := range requiredOwners {
		owner, found := owners[name]
		if !found || nilOwner(owner) {
			return nil, fmt.Errorf("Catalog component owner %s is unavailable", name)
		}
		copyOwners[name] = owner
	}
	for name := range owners {
		if _, found := copyOwners[name]; !found {
			return nil, fmt.Errorf("Catalog component owner %s is unsupported", name)
		}
	}
	return &StaticDirectory{owners: copyOwners}, nil
}

func (directory *StaticDirectory) Resolve(owner string) (Owner, error) {
	if directory == nil || directory.owners == nil {
		return nil, fmt.Errorf("Catalog component owner directory is unavailable")
	}
	resolved, found := directory.owners[owner]
	if !found || nilOwner(resolved) {
		return nil, fmt.Errorf("Catalog component owner %s is unavailable", owner)
	}
	return resolved, nil
}

func nilOwner(owner Owner) bool {
	if owner == nil {
		return true
	}
	value := reflect.ValueOf(owner)
	return value.Kind() == reflect.Pointer && value.IsNil()
}

var _ Directory = (*StaticDirectory)(nil)
