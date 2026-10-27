package catalogprojection

import (
	"crypto/sha256"
	"encoding/json"
)

func projectionDigest(value verifiedRequest) ([32]byte, error) {
	documentSHA256 := sha256.Sum256(value.request.Document)
	canonical, err := json.Marshal(struct {
		PackageID, PackageSlug, PackageVersion, ReleaseID string
		ComponentID, Owner, Contract                      string
		Generation                                        int64
		ContentSHA256, IdempotencyKey                     string
		Kind, ModuleSlug, Identity                        string
		SemanticSHA256, DocumentSHA256                    [32]byte
		ItemCount, AuxiliaryCount                         int
	}{
		value.request.PackageID, value.request.PackageSlug,
		value.request.PackageVersion, value.request.ReleaseID,
		value.request.Component.ID, value.request.Component.Owner,
		value.request.Component.Contract, value.request.Component.Generation,
		value.request.Component.ContentSHA256, value.request.IdempotencyKey,
		value.materialization.Kind, value.materialization.ModuleSlug,
		value.materialization.Identity,
		value.materialization.SemanticSHA256, documentSHA256,
		value.materialization.ItemCount, value.materialization.AuxiliaryCount,
	})
	if err != nil {
		return [32]byte{}, ErrInvalid
	}
	return sha256.Sum256(canonical), nil
}
