package recoverys3

import (
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func privateACL(output *s3.GetBucketAclOutput, expectedOwner string) bool {
	if output == nil || output.Owner == nil || output.Owner.ID == nil ||
		*output.Owner.ID != expectedOwner {
		return false
	}
	for _, grant := range output.Grants {
		if grant.Grantee == nil || grant.Grantee.Type != types.TypeCanonicalUser ||
			grant.Grantee.ID == nil || *grant.Grantee.ID != expectedOwner ||
			grant.Grantee.URI != nil || grant.Grantee.EmailAddress != nil {
			return false
		}
	}
	return true
}
