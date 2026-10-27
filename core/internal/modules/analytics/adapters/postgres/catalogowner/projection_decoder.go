package catalogowner

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	catalogprojection "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogprojection/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/analytics/api/catalogpayload"
)

type projectionDecoder struct{}

func (projectionDecoder) Decode(
	document []byte,
	component contract.Component,
) (catalogprojection.Materialization, error) {
	verified, err := catalogpayload.Verify(
		document, component.ID, component.ContentSHA256,
	)
	if err != nil {
		return catalogprojection.Materialization{}, err
	}
	payload := verified.Value()
	moduleSlug, err := catalogpackagev1.ParseFirstPartySourceContract(
		payload.SourceContract,
	)
	if err != nil {
		return catalogprojection.Materialization{}, err
	}
	digest, err := hex.DecodeString(payload.LegacyHash)
	if err != nil || len(digest) != 32 {
		return catalogprojection.Materialization{}, fmt.Errorf(
			"analytics semantic digest is invalid",
		)
	}
	var definition struct {
		DataSources []json.RawMessage `json:"dataSources"`
		Fields      []json.RawMessage `json:"fields"`
		Widgets     []json.RawMessage `json:"widgets"`
	}
	if err = json.Unmarshal(payload.Definition, &definition); err != nil {
		return catalogprojection.Materialization{}, err
	}
	var semanticSHA256 [32]byte
	copy(semanticSHA256[:], digest)
	return catalogprojection.Materialization{
		Kind: strings.ToUpper(payload.Kind), ModuleSlug: moduleSlug,
		Identity: payload.Slug,
		Document: verified.CanonicalJSON(), SemanticSHA256: semanticSHA256,
		ItemCount: 1,
		AuxiliaryCount: len(definition.DataSources) +
			len(definition.Fields) + len(definition.Widgets),
	}, nil
}

var _ catalogprojection.Decoder = projectionDecoder{}
