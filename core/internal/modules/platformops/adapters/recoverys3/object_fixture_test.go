package recoverys3

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type storedObject struct {
	body    []byte
	version string
}

type objectAPIFixture struct {
	objects  map[string]storedObject
	gets     int
	puts     int
	retained time.Time
}

func objectKey(bucket, key, version string) string {
	return bucket + "\x00" + key + "\x00" + version
}

func (fixture *objectAPIFixture) GetObject(
	_ context.Context,
	input *s3.GetObjectInput,
	_ ...func(*s3.Options),
) (*s3.GetObjectOutput, error) {
	fixture.gets++
	stored, found := fixture.objects[objectKey(
		aws.ToString(input.Bucket),
		aws.ToString(input.Key),
		aws.ToString(input.VersionId),
	)]
	if !found {
		return nil, ErrObject
	}
	return &s3.GetObjectOutput{
		Body: io.NopCloser(bytes.NewReader(stored.body)),
	}, nil
}

func (fixture *objectAPIFixture) PutObject(
	_ context.Context,
	input *s3.PutObjectInput,
	_ ...func(*s3.Options),
) (*s3.PutObjectOutput, error) {
	fixture.puts++
	body, err := io.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	version := "version-" + string(rune('0'+fixture.puts))
	if input.ObjectLockRetainUntilDate != nil {
		fixture.retained = input.ObjectLockRetainUntilDate.UTC()
	}
	fixture.objects[objectKey(
		aws.ToString(input.Bucket), aws.ToString(input.Key), version,
	)] = storedObject{body: body, version: version}
	return &s3.PutObjectOutput{VersionId: aws.String(version)}, nil
}

func (fixture *objectAPIFixture) GetObjectRetention(
	context.Context,
	*s3.GetObjectRetentionInput,
	...func(*s3.Options),
) (*s3.GetObjectRetentionOutput, error) {
	return &s3.GetObjectRetentionOutput{
		Retention: &types.ObjectLockRetention{
			Mode:            types.ObjectLockRetentionModeCompliance,
			RetainUntilDate: &fixture.retained,
		},
	}, nil
}

func newObjectAPIFixture() *objectAPIFixture {
	return &objectAPIFixture{
		objects:  make(map[string]storedObject),
		retained: time.Now().UTC().Add(30 * 24 * time.Hour),
	}
}
