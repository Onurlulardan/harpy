package systemtransfer

import (
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestCompile(t *testing.T) {
	policy, err := newMappingPolicy(publicresource.Users, nil)
	if err != nil {
		t.Fatal(err)
	}
	missingRole := []byte(`{"contract":"tabular-mapping.v2","fields":[` +
		`{"target":"` + publicresource.UserEmailFieldID + `","required":true,` +
		`"expression":{"kind":"direct","source":"email"}}]}`)
	if _, err = policy.Compile(missingRole); !errors.Is(err, ErrInviteRequired) {
		t.Fatalf("missing invite role error=%v", err)
	}
	document := []byte(`{"contract":"tabular-mapping.v2","fields":[` +
		`{"target":"` + publicresource.UserEmailFieldID + `","required":true,` +
		`"expression":{"kind":"direct","source":"email"}},` +
		`{"target":"` + publicresource.UserRoleIDsFieldID + `","required":true,` +
		`"expression":{"kind":"direct","source":"role_ids"}}]}`)
	if _, err = policy.Compile(document); err != nil {
		t.Fatalf("invite mapping rejected: %v", err)
	}
}

func TestMappingRejectsCredentialsAndProtectedFields(t *testing.T) {
	policy, err := newMappingPolicy(publicresource.Users, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, target := range []string{"password", publicresource.UserStatusFieldID,
		publicresource.UserRevisionFieldID} {
		document := []byte(`{"contract":"tabular-mapping.v2","fields":[` +
			`{"target":"` + target + `","required":true,` +
			`"expression":{"kind":"static","value":"secret"}}]}`)
		if _, err = policy.Compile(document); !errors.Is(err, ErrForbiddenTarget) {
			t.Fatalf("target %s error=%v", target, err)
		}
	}
}

func TestIdentityManifestIsImmutableAndResourceSpecific(t *testing.T) {
	digests := make(map[[32]byte]struct{})
	for _, resource := range []publicresource.Resource{publicresource.Users,
		publicresource.Roles, publicresource.Organizations} {
		manifest, err := ManifestFor(resource)
		if err != nil || len(manifest.Fields) == 0 || manifest.Digest == [32]byte{} {
			t.Fatalf("resource=%s manifest=%+v err=%v", resource, manifest, err)
		}
		digests[manifest.Digest] = struct{}{}
	}
	if len(digests) != 3 {
		t.Fatal("resource identity manifests share a digest")
	}
}
