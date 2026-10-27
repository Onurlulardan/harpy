package catalogowner

import (
	"crypto/sha256"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogprojection "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogprojection/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/api/catalogpayload"
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
	return catalogprojection.Materialization{
		Kind: "IAM_TEMPLATE", ModuleSlug: payload.ModuleSlug,
		Identity:       payload.ModuleSlug,
		Document:       verified.CanonicalJSON(),
		SemanticSHA256: sha256.Sum256(document),
		ItemCount:      len(payload.Resources),
		AuxiliaryCount: len(payload.Grants),
	}, nil
}

var _ catalogprojection.Decoder = projectionDecoder{}
