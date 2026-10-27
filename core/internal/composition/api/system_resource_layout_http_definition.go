package api

import (
	"net/http"

	layouts "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresourcelayouts"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpinput"
)

const layoutHTTPDefinitionLimit int64 = 1 << 20

type layoutHTTPDefinition struct {
	Contract layouts.DefinitionContract  `json:"contract"`
	Surfaces []layouts.SurfaceDefinition `json:"surfaces"`
}

type layoutHTTPDefinitionRequest struct {
	Definition layoutHTTPDefinition `json:"definition"`
}

func decodeLayoutHTTPDefinition(
	w http.ResponseWriter, r *http.Request,
) (layouts.Definition, error) {
	var document layoutHTTPDefinitionRequest
	if err := httpinput.DecodeJSON(
		w, r, layoutHTTPDefinitionLimit, &document,
	); err != nil {
		return layouts.Definition{}, err
	}
	if document.Definition.Contract != layouts.DefinitionContractV2 {
		return layouts.Definition{}, layouts.ErrLayoutInvalid
	}
	return layouts.Definition{
		SourceContract: document.Definition.Contract,
		Surfaces:       document.Definition.Surfaces,
	}, nil
}
