package systemresources

import publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"

const AuthorityGroup = publicresource.AuthorityGroup

type Admission = publicresource.Admission

func NewAdmission(
	tenantID publicresource.ID,
	actor publicresource.ActorEvidence,
	groupGeneration int64,
	iamGeneration int64,
	layoutGeneration int64,
) (Admission, error) {
	admission, err := publicresource.NewAdmission(
		tenantID, actor, groupGeneration, iamGeneration, layoutGeneration,
	)
	return admission, normalizeError(err)
}
