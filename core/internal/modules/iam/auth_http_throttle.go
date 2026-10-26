package iam

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"slices"
	"strings"

	"github.com/google/uuid"
)

type authHTTPThrottleBucket struct {
	scopeKind, purpose, dimension  string
	tenantID                       uuid.UUID
	keyID                          string
	hmac                           [32]byte
	groupGeneration, iamGeneration int64
}

func (application *authHTTPApplication) throttleBucket(
	scopeKind, purpose, dimension string,
	tenantID uuid.UUID, groupGeneration, iamGeneration int64,
	values ...string,
) (authHTTPThrottleBucket, error) {
	if len(values) == 0 {
		return authHTTPThrottleBucket{}, ErrAuthHTTPUnavailable
	}
	keyID, key, err := application.throttle.Active()
	if err != nil {
		return authHTTPThrottleBucket{}, ErrAuthHTTPUnavailable
	}
	mac := hmac.New(sha256.New, key[:])
	clear(key[:])
	parts := append(
		[]string{"harpy:iam:auth-throttle:v1", scopeKind, purpose, dimension},
		values...,
	)
	for _, part := range parts {
		if part == "" {
			return authHTTPThrottleBucket{}, ErrAuthHTTPUnavailable
		}
		var size [4]byte
		binary.BigEndian.PutUint32(size[:], uint32(len(part)))
		_, _ = mac.Write(size[:])
		_, _ = mac.Write([]byte(part))
	}
	var digest [32]byte
	copy(digest[:], mac.Sum(nil))
	return authHTTPThrottleBucket{
		scopeKind: scopeKind, purpose: purpose, dimension: dimension,
		tenantID: tenantID, keyID: keyID, hmac: digest,
		groupGeneration: groupGeneration, iamGeneration: iamGeneration,
	}, nil
}

func (application *authHTTPApplication) throttleBuckets(
	scopeKind, purpose string,
	tenantID uuid.UUID,
	groupGeneration, iamGeneration int64,
	dimensions map[string][]string,
) ([]authHTTPThrottleBucket, error) {
	buckets := make([]authHTTPThrottleBucket, 0, len(dimensions))
	for dimension, values := range dimensions {
		bucket, err := application.throttleBucket(
			scopeKind, purpose, dimension, tenantID,
			groupGeneration, iamGeneration, values...,
		)
		if err != nil {
			return nil, err
		}
		buckets = append(buckets, bucket)
	}
	slices.SortFunc(buckets, func(left, right authHTTPThrottleBucket) int {
		return strings.Compare(left.dimension, right.dimension)
	})
	return buckets, nil
}

func (application *authHTTPApplication) throttleAnyBlocked(
	ctx context.Context, buckets []authHTTPThrottleBucket,
) error {
	for _, bucket := range buckets {
		if err := application.throttleBlocked(ctx, bucket); err != nil {
			return err
		}
	}
	return nil
}

func (application *authHTTPApplication) throttleBlocked(
	ctx context.Context, bucket authHTTPThrottleBucket,
) error {
	blocked, err := application.pool.IAMThrottleBlocked(
		ctx, bucket.scopeKind, bucket.tenantID, bucket.purpose, bucket.dimension,
		bucket.keyID, bucket.hmac, bucket.groupGeneration, bucket.iamGeneration,
	)
	clear(bucket.hmac[:])
	if err != nil {
		return ErrAuthHTTPUnavailable
	}
	if blocked {
		return ErrAuthHTTPInvalidCredentials
	}
	return nil
}

func (application *authHTTPApplication) recordThrottleFailure(
	ctx context.Context, buckets ...authHTTPThrottleBucket,
) error {
	for _, bucket := range buckets {
		_, err := application.pool.IAMThrottleRecordFailure(
			ctx, bucket.scopeKind, bucket.tenantID, bucket.purpose, bucket.dimension,
			bucket.keyID, bucket.hmac, bucket.groupGeneration, bucket.iamGeneration,
		)
		clear(bucket.hmac[:])
		if err != nil {
			return ErrAuthHTTPUnavailable
		}
	}
	return nil
}
