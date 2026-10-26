package documents

import (
	"fmt"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
)

// Production exposes only stable owner/application interfaces. External clients
// remain private to the immutable Documents production graph.
type Production struct {
	application documentshttp.Application
	knowledge   knowledgeapi.FileOwner
}

func newProduction(application documentshttp.Application,
	knowledge knowledgeapi.FileOwner,
) (*Production, error) {
	if nilHTTPPort(application) || nilHTTPPort(knowledge) {
		return nil, fmt.Errorf("Documents production owners are unavailable")
	}
	return &Production{application: application, knowledge: knowledge}, nil
}

func (production *Production) Application() documentshttp.Application {
	if production == nil || nilHTTPPort(production.application) {
		return nil
	}
	return production.application
}

func (production *Production) KnowledgeFileOwner() knowledgeapi.FileOwner {
	if production == nil || nilHTTPPort(production.knowledge) {
		return nil
	}
	return production.knowledge
}
