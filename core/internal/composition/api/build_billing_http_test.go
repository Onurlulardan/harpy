package api

import (
	"testing"

	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func TestBuildBillingHTTPFailsClosedOnIncompleteComposition(t *testing.T) {
	_, err := buildBillingHTTP(
		config.API{}, nil, platformasync.WriteCatalogs{}, nil, func(error) {},
	)
	if err == nil {
		t.Fatal("incomplete Billing HTTP composition was accepted")
	}
}
