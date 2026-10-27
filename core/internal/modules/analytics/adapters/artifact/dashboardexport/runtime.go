package dashboardexportartifact

import (
	"fmt"
	"regexp"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
)

var prefixPattern = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)

type Runtime struct {
	spool   *spoolapi.Manager
	objects artifactobject.Store
	keys    *fcartv1.EnvelopeKeyring
	prefix  string
}

func NewRuntime(
	spool *spoolapi.Manager, objects artifactobject.Store,
	keys *fcartv1.EnvelopeKeyring, prefix string,
) (*Runtime, error) {
	if spool == nil || objects == nil || keys == nil || !prefixPattern.MatchString(prefix) {
		return nil, fmt.Errorf("analytics dashboard export artifact dependencies are invalid")
	}
	return &Runtime{spool: spool, objects: objects, keys: keys, prefix: prefix}, nil
}

func NewContentRuntime(
	objects artifactobject.Store, keys *fcartv1.EnvelopeKeyring, prefix string,
) (exportapp.ContentRuntime, error) {
	if objects == nil || keys == nil || !prefixPattern.MatchString(prefix) {
		return nil, fmt.Errorf("analytics dashboard export content dependencies are invalid")
	}
	return &Runtime{objects: objects, keys: keys, prefix: prefix}, nil
}

var _ exportapp.ArtifactRuntime = (*Runtime)(nil)
