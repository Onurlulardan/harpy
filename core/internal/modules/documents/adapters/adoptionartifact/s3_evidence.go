package adoptionartifact

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/adoption"
)

const (
	metadataState       = "harpy-adoption-state"
	metadataRun         = "harpy-adoption-run-id"
	metadataDocument    = "harpy-source-document-id"
	metadataVersion     = "harpy-source-version-id"
	metadataLSN         = "harpy-source-lsn"
	metadataContentHash = "harpy-source-content-sha256"
	metadataRecordHash  = "harpy-source-metadata-sha256"
)

func (source *S3Source) VerifyInactive(ctx context.Context,
	checkpoint adoptionapp.Checkpoint,
) ([32]byte, error) {
	if source == nil || ctx == nil || ctx.Err() != nil ||
		!validEvidenceCheckpoint(checkpoint) {
		return [32]byte{}, adoptionapp.ErrInvalid
	}
	output, err := source.api.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: &source.bucket, Key: &checkpoint.SourceObjectKey,
		VersionId: &checkpoint.SourceRevision, ExpectedBucketOwner: source.owner})
	if err != nil {
		return [32]byte{}, errors.Join(adoptionapp.ErrUnavailable, err)
	}
	if output == nil || output.VersionId == nil ||
		*output.VersionId != checkpoint.SourceRevision || output.ContentLength == nil ||
		*output.ContentLength != checkpoint.SourceBytes ||
		!matchingEvidence(output.Metadata, checkpoint) {
		return [32]byte{}, adoptionapp.ErrBlocked
	}
	decoded, err := hex.DecodeString(output.Metadata[metadataRecordHash])
	if err != nil || len(decoded) != 32 ||
		hex.EncodeToString(decoded) != output.Metadata[metadataRecordHash] {
		clear(decoded)
		return [32]byte{}, adoptionapp.ErrBlocked
	}
	var result [32]byte
	copy(result[:], decoded)
	clear(decoded)
	return result, nil
}

func validEvidenceCheckpoint(value adoptionapp.Checkpoint) bool {
	return value.RunID.UUIDv7() && value.SourceItemID.UUIDv7() &&
		value.SourceVersionID.UUIDv7() && validSourceName(value.SourceObjectKey) &&
		validSourceName(value.SourceRevision) && value.SourceBytes > 0 &&
		value.SourceSHA256 != [32]byte{} && value.SourceMetadataSHA256 != [32]byte{} &&
		value.SourceLSN != ""
}

func matchingEvidence(metadata map[string]string,
	checkpoint adoptionapp.Checkpoint,
) bool {
	return metadata[metadataState] == "inactive" &&
		metadata[metadataRun] == checkpoint.RunID.String() &&
		metadata[metadataDocument] == checkpoint.SourceItemID.String() &&
		metadata[metadataVersion] == checkpoint.SourceVersionID.String() &&
		metadata[metadataLSN] == checkpoint.SourceLSN &&
		metadata[metadataContentHash] == hex.EncodeToString(checkpoint.SourceSHA256[:]) &&
		metadata[metadataRecordHash] == hex.EncodeToString(checkpoint.SourceMetadataSHA256[:])
}

var _ MetadataVerifier = (*S3Source)(nil)
