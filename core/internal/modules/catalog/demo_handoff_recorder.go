package catalog

import (
	catalogownerpostgres "github.com/Onurlulardan/harpy/core/internal/modules/catalog/adapters/postgres/catalogowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/demohandoff"
	handoffapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/demohandoff"
)

func NewDemoHandoffRecorder() (demohandoff.Recorder, error) {
	values := catalogComponentValues{}
	return handoffapp.New(catalogownerpostgres.NewDemoHandoffBinder(), values)
}
