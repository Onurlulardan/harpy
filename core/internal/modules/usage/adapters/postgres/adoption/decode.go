package adoption

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"

	app "github.com/Onurlulardan/harpy/core/internal/modules/usage/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func strictDecode(document []byte, target any) error {
	if len(document) == 0 || len(document) > 1024*1024 {
		return domain.ErrUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(target) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return domain.ErrUnavailable
	}
	return nil
}

func decodeRun(document []byte) (app.Run, error) {
	var wire runDocument
	if strictDecode(document, &wire) != nil {
		return app.Run{}, domain.ErrUnavailable
	}
	root, err := digest(wire.SourceRootSHA256)
	if err != nil {
		return app.Run{}, err
	}
	transform, err := digest(wire.TransformSHA256)
	if err != nil {
		return app.Run{}, err
	}
	run := app.Run{RunID: wire.RunID, OperationID: wire.OperationID,
		Manifest: app.Manifest{Kind: wire.SourceKind, Schema: wire.SourceSchema,
			Relation: wire.SourceRelation, RelationOID: wire.SourceRelationOID,
			RowCount: wire.SourceCount, RootSHA256: root, SnapshotID: wire.SnapshotID},
		TransformSHA256: transform, NextPage: wire.NextPage, CursorID: wire.CursorID,
		CopiedCount: wire.CopiedCount, Status: wire.Status, LeaseOwner: wire.LeaseOwner,
		LeaseGeneration: wire.LeaseGeneration, LeaseExpiresAt: wire.LeaseExpiresAt,
		Revision: wire.Revision, CompletedAt: wire.CompletedAt}
	if run.Validate() != nil {
		return app.Run{}, domain.ErrUnavailable
	}
	return run, nil
}

func decodeReceipt(document []byte) (app.PageReceipt, error) {
	var wire receiptDocument
	if strictDecode(document, &wire) != nil || wire.RunID == [16]byte{} ||
		wire.PageNumber < 1 || wire.CursorAfter == [16]byte{} || wire.RowCount < 1 ||
		wire.RowCount > 100 || wire.State != "applied" && wire.State != "advanced" ||
		wire.AppliedAt.IsZero() {
		return app.PageReceipt{}, domain.ErrUnavailable
	}
	return app.PageReceipt{PageNumber: wire.PageNumber,
		RowCount: wire.RowCount, State: wire.State}, nil
}

func digest(value string) ([32]byte, error) {
	decoded, err := hex.DecodeString(value)
	if err != nil || len(decoded) != 32 {
		return [32]byte{}, domain.ErrUnavailable
	}
	var result [32]byte
	copy(result[:], decoded)
	return result, nil
}
