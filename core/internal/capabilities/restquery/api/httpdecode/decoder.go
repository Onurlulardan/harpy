package httpdecode

import (
	transport "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/http"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/published"
)

// GET returns the published REST Query GET transport decoder.
func GET() api.GETDecoder { return transport.NewGETDecoder() }

// POST returns the published REST Query POST transport decoder.
func POST() api.POSTDecoder { return transport.NewPOSTDecoder() }

// PublishedGET returns the final typed parameter decoder for published reads.
func PublishedGET() published.GETDecoder { return transport.NewPublishedGETDecoder() }

// PublishedPOST returns the final typed parameter decoder for published reads.
func PublishedPOST() published.POSTDecoder { return transport.NewPublishedPOSTDecoder() }
