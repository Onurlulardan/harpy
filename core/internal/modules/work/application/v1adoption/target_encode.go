package v1adoption

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

var adoptionNamespace = uuid.MustParse("f32f48db-43df-52b6-8af8-281c9d95c047")

func targetRecord(source SourceRecord, kind TargetKind, id domain.ID, value any) (TargetRecord, error) {
	document, err := json.Marshal(value)
	if err != nil || !id.Valid() {
		return TargetRecord{}, ports.ErrUnavailable
	}
	return TargetRecord{SourceKind: source.Kind, SourceID: source.ID, Kind: kind,
		ID: id, Document: document, SourceSHA256: source.SHA256,
		SHA256: sha256.Sum256(document)}, nil
}

func derivedID(parts ...string) domain.ID {
	value := ""
	for _, part := range parts {
		value += "\x00" + part
	}
	return domain.ID(uuid.NewSHA1(adoptionNamespace, []byte(value)))
}

func adoptionIssue(source SourceRecord, code IssueCode) Issue {
	hash := sha256.New()
	_, _ = hash.Write([]byte(code))
	_, _ = hash.Write(source.SHA256[:])
	var evidence [sha256.Size]byte
	copy(evidence[:], hash.Sum(nil))
	return Issue{Code: code, SourceKind: source.Kind,
		SourceID: source.ID, Evidence: evidence}
}

func appendTarget(plan *Plan, source SourceRecord, kind TargetKind, id domain.ID, value any) error {
	record, err := targetRecord(source, kind, id, value)
	if err != nil {
		return err
	}
	plan.Records = append(plan.Records, record)
	return nil
}

func appendIssue(plan *Plan, source SourceRecord, code IssueCode) {
	for _, issue := range plan.Issues {
		if issue.SourceKind == source.Kind && issue.SourceID == source.ID && issue.Code == code {
			return
		}
	}
	plan.Issues = append(plan.Issues, adoptionIssue(source, code))
}
