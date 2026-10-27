package iam

import (
	"fmt"
	"net/http"

	postrestorepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/postrestore"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/password"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/httpedge"
)

type PostRestoreRuntime struct {
	http      *PostRestoreHTTP
	admission *PostRestoreAdmission
}

func NewPostRestoreRuntime(
	pool postrestorepostgres.RuntimePool, settings config.IAMAPI,
	common, compromised IdentityPasswordRejector, auth *AuthHTTP, report HTTPErrorSink,
) (*PostRestoreRuntime, error) {
	store, err := postrestorepostgres.NewStore(pool)
	if err != nil {
		return nil, fmt.Errorf("build IAM post-restore store: %w", err)
	}
	policy, err := password.NewPolicy(common, compromised)
	if err != nil {
		return nil, fmt.Errorf("build IAM post-restore password policy: %w", err)
	}
	directory, err := activePostRestoreDirectoryKey(settings)
	if err != nil {
		return nil, err
	}
	profile, err := settings.HTTPProfile()
	if err != nil {
		return nil, fmt.Errorf("load IAM post-restore HTTP profile: %w", err)
	}
	security := postrestore.Security{
		Challenges: postRestoreChallengeGenerator{},
		Assertions: postRestoreAssertionVerifier{},
		Passwords:  postRestorePasswordIssuer{policy: policy},
	}
	service, err := postrestore.NewService(store, security, directory,
		profile.AllowedOrigins(), postrestore.SystemClock, postrestore.SystemUUID)
	if err != nil {
		return nil, fmt.Errorf("build IAM post-restore service: %w", err)
	}
	handler, err := newPostRestoreHTTP(service, auth, report)
	if err != nil {
		return nil, err
	}
	admission, err := NewPostRestoreAdmission(store, report)
	if err != nil {
		return nil, err
	}
	return &PostRestoreRuntime{http: handler, admission: admission}, nil
}

func activePostRestoreDirectoryKey(settings config.IAMAPI) (postrestore.DirectoryKey, error) {
	ring, err := settings.LoginDirectoryKeyring()
	if err != nil {
		return postrestore.DirectoryKey{}, err
	}
	keys, err := ring.Keys()
	if err != nil || len(keys) == 0 {
		return postrestore.DirectoryKey{}, fmt.Errorf("IAM post-restore directory key is unavailable")
	}
	id, idErr := keys[0].ID()
	material, materialErr := keys[0].Material()
	active, activeErr := keys[0].Active()
	for index := range keys {
		keys[index].Destroy()
	}
	if idErr != nil || materialErr != nil || activeErr != nil || !active {
		clear(material[:])
		return postrestore.DirectoryKey{}, fmt.Errorf("IAM post-restore directory key is invalid")
	}
	return postrestore.DirectoryKey{ID: id, Material: material}, nil
}

func (runtime *PostRestoreRuntime) Routes() []httpedge.Route {
	if runtime == nil || runtime.http == nil {
		return nil
	}
	return runtime.http.Routes()
}

func (runtime *PostRestoreRuntime) Wrap(next http.Handler) (http.Handler, error) {
	if runtime == nil || runtime.admission == nil {
		return nil, fmt.Errorf("IAM post-restore admission is unavailable")
	}
	return runtime.admission.Wrap(next)
}
