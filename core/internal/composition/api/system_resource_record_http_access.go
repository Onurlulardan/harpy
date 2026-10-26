package api

import (
	"net/http"

	publicresource "github.com/Onurlulardan/harpy/core/internal/orchestration/systemresources"
)

func (handler *systemResourceRecordHTTPHandler) recordAdmission(
	r *http.Request,
	resource publicresource.Resource,
	action string,
) (publicresource.Admission, error) {
	operation, valid := systemResourceRecordOperation(resource, action)
	if !valid {
		return publicresource.Admission{}, publicresource.ErrLayoutAccessAuthentication
	}
	return handler.recordAdmissionForOperation(r, resource, action, operation)
}

func (handler *systemResourceRecordHTTPHandler) recordReadAdmission(
	r *http.Request,
	resource publicresource.Resource,
) (publicresource.Admission, error) {
	operation, valid := systemResourceRecordReadOperation(resource)
	if !valid {
		return publicresource.Admission{}, publicresource.ErrLayoutAccessAuthentication
	}
	return handler.recordAdmissionForOperation(
		r, resource, systemResourceRecordHTTPRead, operation,
	)
}

func (handler *systemResourceRecordHTTPHandler) recordAdmissionForOperation(
	r *http.Request,
	resource publicresource.Resource,
	action, operation string,
) (publicresource.Admission, error) {
	if handler == nil || r == nil {
		return publicresource.Admission{}, publicresource.ErrLayoutAccessAuthentication
	}
	decision, err := handler.access.Verify(
		r.Context(),
		publicresource.LayoutAccessRequest{
			AuthorizationValues: append(
				[]string(nil), r.Header.Values("Authorization")...,
			),
			TenantIDValues: append(
				[]string(nil), r.Header.Values("X-Tenant-ID")...,
			),
			TenantSlugValues: append(
				[]string(nil), r.Header.Values("X-Tenant-Slug")...,
			),
			Resource: resource, Action: action, Operation: operation,
		},
	)
	if err != nil {
		return publicresource.Admission{}, err
	}
	admission, err := decision.Admission()
	if err != nil {
		return publicresource.Admission{}, publicresource.ErrLayoutAccessUnavailable
	}
	return admission, nil
}

func systemResourceRecordOperation(
	resource publicresource.Resource,
	action string,
) (string, bool) {
	name := map[publicresource.Resource]string{
		publicresource.Users:         "user",
		publicresource.Roles:         "role",
		publicresource.Organizations: "organization",
	}[resource]
	if name == "" {
		return "", false
	}
	suffix := action
	if action == systemResourceRecordHTTPRead {
		suffix = "list"
	}
	if action == systemResourceRecordHTTPDelete && resource == publicresource.Roles {
		suffix = "retire"
	}
	return "iam.identity." + name + "." + suffix, true
}

func systemResourceRecordReadOperation(
	resource publicresource.Resource,
) (string, bool) {
	name := map[publicresource.Resource]string{
		publicresource.Users:         "user",
		publicresource.Roles:         "role",
		publicresource.Organizations: "organization",
	}[resource]
	return "iam.identity." + name + ".read", name != ""
}
