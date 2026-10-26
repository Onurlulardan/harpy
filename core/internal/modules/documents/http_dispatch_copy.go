package documents

import (
	"context"
	"encoding/hex"
	"net/http"

	documentshttp "github.com/Onurlulardan/harpy/core/internal/modules/documents/adapters/http"
	copyapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/copy"
)

type httpCopyItemRequest struct {
	SourceVersionID string   `json:"source_version_id"`
	TargetFolderID  string   `json:"target_folder_id"`
	OriginalName    string   `json:"original_name"`
	Description     string   `json:"description"`
	Tags            []string `json:"tags"`
}

func (runtime *HTTPRuntime) dispatchCopy(ctx context.Context,
	request documentshttp.Request,
) (documentshttp.Response, error) {
	input, err := decodeHTTPRequest[httpCopyItemRequest](request.Document)
	if err != nil || !onlyHTTPQuery(request.Query) {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	itemID, itemErr := parseHTTPID(request.ResourceID, false)
	versionID, versionErr := parseHTTPID(input.SourceVersionID, true)
	folderID, folderErr := parseHTTPID(input.TargetFolderID, false)
	if itemErr != nil || versionErr != nil || folderErr != nil {
		return documentshttp.Response{}, documentshttp.ErrInvalid
	}
	runner := httpScopeRunner{runtime: runtime, access: request.Access}
	result, err := runtime.dependencies.Copies.Copy(ctx, runner, copyapp.Request{
		SourceItemID: itemID, SourceVersionID: versionID, TargetFolderID: folderID,
		OriginalName: input.OriginalName, Description: input.Description,
		Tags: input.Tags, IdempotencyKey: request.IdempotencyKey,
	})
	data := map[string]any{"source_document_id": result.SourceItemID.String(),
		"source_version_id": result.SourceVersionID.String(),
		"document_id":       result.ItemID.String(), "version_id": result.VersionID.String(),
		"upload_id": result.UploadID.String(), "bytes": result.Bytes,
		"sha256": hex.EncodeToString(result.SHA256[:])}
	return documentshttp.Response{Status: http.StatusCreated, Data: data,
			Location: "/api/documents/" + result.ItemID.String()},
		mapHTTPApplicationError(err)
}
