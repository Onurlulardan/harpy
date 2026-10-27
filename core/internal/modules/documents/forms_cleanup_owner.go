package documents

import (
	formscleanuppostgres "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/postgres/formscleanup"
	formscleanupapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/formscleanup"
	formscleanupapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/formscleanup"
)

func NewFormsCleanupOwner(jobs *APIAsyncPort) (formscleanupapi.Owner, error) {
	return formscleanupapp.NewApplication(formscleanuppostgres.NewBinder(),
		jobs, platformIDs{}, platformClock{})
}
