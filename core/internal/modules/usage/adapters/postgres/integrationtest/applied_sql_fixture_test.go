//go:build integration

package integrationtest

type appliedSQL struct {
	scope string
	path  string
	body  []byte
}

func (document appliedSQL) Scope() string { return document.scope }

func (document appliedSQL) Path() string { return document.path }

func (document appliedSQL) Content() []byte { return document.body }

func newAppliedSQL(_, scope, path string, body []byte) (appliedSQL, error) {
	return appliedSQL{scope: scope, path: path, body: body}, nil
}
