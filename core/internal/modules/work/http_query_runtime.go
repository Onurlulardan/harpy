package work

import (
	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	executionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres/restquery/execution"
	storagecatalog "github.com/Onurlulardan/harpy/core/internal/modules/work/adapters/postgres/storagecatalog"
	workquery "github.com/Onurlulardan/harpy/core/internal/modules/work/application/restquery"
)

type RESTQueryCursorKeys = workquery.CursorKeys

func NewRESTQueryCursorKeys(currentID string, current [32]byte,
	previousID string, previous [32]byte,
) (RESTQueryCursorKeys, error) {
	return workquery.NewCursorKeys(currentID, current, previousID, previous)
}

func newWorkQueryRuntime(principals workprincipalapi.Directory,
	keys RESTQueryCursorKeys,
) (*workquery.Service, error) {
	owner, err := executionpostgres.New()
	if err != nil {
		return nil, err
	}
	catalogs, err := storagecatalog.New()
	if err != nil {
		return nil, err
	}
	return workquery.New(principals, catalogs, owner, keys)
}
