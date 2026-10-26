package iam

import (
	"context"
	"fmt"
	"io"
	"os"

	postrestorev1 "github.com/Onurlulardan/harpy/contracts/postrestorereactivation/v1"
	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/postrestore"
	iamapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api"
	"github.com/google/uuid"
)

type postRestoreObserverFacade struct {
	observer *iampostgres.Observer
}

func NewPostRestoreObserver(
	database any,
) (iamapi.PostRestoreObserver, error) {
	observer, err := iampostgres.BindObserver(database)
	if err != nil {
		return nil, err
	}
	return &postRestoreObserverFacade{observer: observer}, nil
}

func ReadPostRestoreAwait(path string) (iamapi.PostRestoreAwait, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() ||
		info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > postrestorev1.MaximumDocumentBytes {
		return iamapi.PostRestoreAwait{},
			fmt.Errorf("post-restore reactivation input is invalid")
	}
	file, err := os.Open(path)
	if err != nil {
		return iamapi.PostRestoreAwait{}, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil || !os.SameFile(info, opened) {
		return iamapi.PostRestoreAwait{},
			fmt.Errorf("post-restore reactivation input changed")
	}
	document, err := io.ReadAll(io.LimitReader(
		file, postrestorev1.MaximumDocumentBytes+1,
	))
	if err != nil || len(document) > postrestorev1.MaximumDocumentBytes {
		return iamapi.PostRestoreAwait{},
			fmt.Errorf("post-restore reactivation input exceeds its limit")
	}
	await, err := postrestorev1.ParseAwait(document)
	return iamapi.PostRestoreAwait{
		EnvironmentID:      await.EnvironmentID,
		RestoreOperationID: await.RestoreOperationID,
	}, err
}

func (facade *postRestoreObserverFacade) ObserveReactivation(
	ctx context.Context,
	environmentID string,
	restoreOperationID string,
) (iamapi.PostRestoreObservation, error) {
	environment, environmentErr := uuid.Parse(environmentID)
	operation, operationErr := uuid.Parse(restoreOperationID)
	if environmentErr != nil || operationErr != nil ||
		environment.String() != environmentID ||
		operation.String() != restoreOperationID {
		return iamapi.PostRestoreObservation{},
			fmt.Errorf("post-restore observation identity is invalid")
	}
	observation, err := facade.observer.ObserveReactivation(
		ctx, environment, operation,
	)
	return iamapi.PostRestoreObservation{
		State: observation.State, ReceiptSHA256: observation.ReceiptSHA256,
		ReactivatedAt: observation.ReactivatedAt,
	}, err
}
