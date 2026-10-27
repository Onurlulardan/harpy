package s3object

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func privateBucketACL(output *s3.GetBucketAclOutput, expectedOwner string) bool {
	if output == nil || output.Owner == nil || output.Owner.ID == nil ||
		*output.Owner.ID == "" {
		return false
	}
	owner := *output.Owner.ID
	if expectedOwner != "" && owner != expectedOwner {
		return false
	}
	for _, grant := range output.Grants {
		if grant.Grantee == nil || grant.Grantee.Type != types.TypeCanonicalUser ||
			grant.Grantee.ID == nil || *grant.Grantee.ID != owner ||
			grant.Grantee.URI != nil || grant.Grantee.EmailAddress != nil {
			return false
		}
	}
	return true
}
