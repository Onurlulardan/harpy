package systemtransferartifact

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"testing"
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
)

var fixtureArtifactTime = time.Date(2026, 7, 18, 12, 0, 0, 0, time.UTC)

type fixtureClock struct{ now time.Time }

func (clock fixtureClock) Now() time.Time { return clock.now }

type fixtureObject struct {
	objects           map[artifactobject.Identity][]byte
	uploads           int
	deletes           int
	deleteErr         error
	beforeRead        func()
	reconcile         artifactobject.ReconcileResult
	reconcileErr      error
	reconcileExpected artifactobject.ReconcileExpected
}

func newFixtureObject() *fixtureObject {
	return &fixtureObject{objects: make(map[artifactobject.Identity][]byte)}
}

func (store *fixtureObject) Upload(_ context.Context,
	request artifactobject.Upload) (artifactobject.UploadResult, error) {
	if store.beforeRead != nil {
		store.beforeRead()
	}
	value, err := io.ReadAll(request.Body)
	if err != nil || int64(len(value)) != request.CiphertextBytes ||
		sha256.Sum256(value) != request.CiphertextSHA256 {
		return artifactobject.UploadResult{}, errors.New("fixture upload mismatch")
	}
	store.uploads++
	identity := artifactobject.Identity{Key: request.Key, Version: "version-1"}
	store.objects[identity] = append([]byte(nil), value...)
	return artifactobject.UploadResult{Identity: identity,
		ProviderRequestID: "provider-upload-1", UploadedAt: fixtureArtifactTime}, nil
}

func (store *fixtureObject) OpenExact(_ context.Context,
	expected artifactobject.Expected) (io.ReadCloser, error) {
	value, found := store.objects[expected.Identity]
	if !found || int64(len(value)) != expected.CiphertextBytes ||
		sha256.Sum256(value) != expected.CiphertextSHA256 {
		return nil, artifactobject.ErrIntegrity
	}
	return io.NopCloser(bytes.NewReader(append([]byte(nil), value...))), nil
}

func (store *fixtureObject) DeleteExact(_ context.Context,
	identity artifactobject.Identity) (artifactobject.DeleteResult, error) {
	if store.deleteErr != nil {
		return artifactobject.DeleteResult{}, store.deleteErr
	}
	delete(store.objects, identity)
	store.deletes++
	return artifactobject.DeleteResult{ProviderRequestID: "provider-delete-1",
		DeleteResponseSHA256: sha256.Sum256([]byte("delete")),
		AbsenceProbeSHA256:   sha256.Sum256([]byte("absent")),
		VerifiedAbsentAt:     fixtureArtifactTime}, nil
}

func (store *fixtureObject) ReconcileExact(_ context.Context,
	expected artifactobject.ReconcileExpected,
) (artifactobject.ReconcileResult, error) {
	store.reconcileExpected = expected
	return store.reconcile, store.reconcileErr
}

func newRuntimeFixture(t *testing.T) (*Runtime, *fixtureObject, *fixtureObject) {
	t.Helper()
	key := bytes.Repeat([]byte{7}, 32)
	importKEK, err := fcartv1.NewKEK("iam-import-kek-1", key)
	if err != nil {
		t.Fatal(err)
	}
	exportKEK, err := fcartv1.NewKEK("iam-export-kek-1", bytes.Repeat([]byte{8}, 32))
	if err != nil {
		t.Fatal(err)
	}
	importKeys, err := fcartv1.NewEnvelopeKeyring(importKEK, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	exportKeys, err := fcartv1.NewEnvelopeKeyring(exportKEK, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	spools, err := spoolapi.NewManager(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	imports, exports := newFixtureObject(), newFixtureObject()
	runtime, err := NewRuntime(imports, exports, importKeys, exportKeys,
		spools, fixtureClock{now: fixtureArtifactTime})
	if err != nil {
		t.Fatal(err)
	}
	return runtime, imports, exports
}
