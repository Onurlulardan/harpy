package catalog

import "github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"

type ComponentOwner = componentowner.Owner
type ComponentOwnerDirectory = componentowner.Directory

func NewComponentOwnerDirectory(
	analytics ComponentOwner,
	catalog ComponentOwner,
	iam ComponentOwner,
) (ComponentOwnerDirectory, error) {
	return componentowner.NewDirectory(map[string]componentowner.Owner{
		"analytics": analytics,
		"catalog":   catalog,
		"iam":       iam,
	})
}
