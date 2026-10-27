//go:build integration

package systemtransferartifact

import (
	"bytes"
	"context"
	"fmt"
	"net/netip"
	"testing"
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactintegration "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api/integration"
	spoolapi "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactspool/api"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/moby/moby/api/types/container"
	mobynetwork "github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const iamMinIOImage = "minio/minio:latest@sha256:" +
	"14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e"

type minIOArtifactClock struct{}

func (minIOArtifactClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

type minIORuntimeFixture struct {
	runtime  *Runtime
	client   *s3.Client
	endpoint string
	imports  string
	exports  string
}

func newMinIORuntimeFixture(t *testing.T) minIORuntimeFixture {
	t.Helper()
	client, endpoint := startIAMMinIO(t)
	imports, exports := "iam-import-integration", "iam-export-integration"
	createIAMMinIOBucket(t, client, imports)
	createIAMMinIOBucket(t, client, exports)
	clock := minIOArtifactClock{}
	importStore := verifiedIAMMinIOStore(t, endpoint, imports,
		ImportObjectPrefix, ImportObjectContract)
	exportStore := verifiedIAMMinIOStore(t, endpoint, exports,
		ExportObjectPrefix, ExportObjectContract)
	importKEK, err := fcartv1.NewKEK("iam-import-integration-kek",
		bytes.Repeat([]byte{7}, 32))
	if err != nil {
		t.Fatal(err)
	}
	exportKEK, err := fcartv1.NewKEK("iam-export-integration-kek",
		bytes.Repeat([]byte{8}, 32))
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
	runtime, err := NewRuntime(importStore, exportStore, importKeys, exportKeys,
		spools, clock)
	if err != nil {
		t.Fatal(err)
	}
	return minIORuntimeFixture{runtime: runtime, client: client, endpoint: endpoint,
		imports: imports, exports: exports}
}

func verifiedIAMMinIOStore(t *testing.T, endpoint, bucket,
	prefix, contract string,
) ObjectStore {
	t.Helper()
	store, err := artifactintegration.OpenStaticS3(t.Context(), endpoint, "us-east-1",
		bucket, "", prefix, contract, true, int64(fcartv1.MaxCiphertextBytes),
		"harpy-integration", "integration-secret-123", "")
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func requireIAMMinIOPublicACLRejected(t *testing.T, fixture minIORuntimeFixture) {
	t.Helper()
	bucket := "iam-public-integration"
	createIAMMinIOBucket(t, fixture.client, bucket)
	_, err := artifactintegration.OpenStaticS3PublicACLProbe(t.Context(),
		fixture.endpoint, "us-east-1",
		bucket, "", ImportObjectPrefix, ImportObjectContract, true,
		int64(fcartv1.MaxCiphertextBytes), "harpy-integration",
		"integration-secret-123", "")
	if err == nil {
		t.Fatal("public MinIO bucket ACL passed artifact store preflight")
	}
}

func createIAMMinIOBucket(t *testing.T, client *s3.Client, bucket string) {
	t.Helper()
	if _, err := client.CreateBucket(t.Context(),
		&s3.CreateBucketInput{Bucket: &bucket}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PutBucketVersioning(t.Context(), &s3.PutBucketVersioningInput{
		Bucket: &bucket, VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled}}); err != nil {
		t.Fatal(err)
	}
}

func startIAMMinIO(t *testing.T) (*s3.Client, string) {
	t.Helper()
	hostPort := reserveIAMMinIOPort(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	instance, err := testcontainers.GenericContainer(ctx,
		testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{
			Image: iamMinIOImage, Cmd: []string{"server", "/data", "--address", ":9000"},
			Env: map[string]string{"MINIO_ROOT_USER": "harpy-integration",
				"MINIO_ROOT_PASSWORD": "integration-secret-123"},
			ExposedPorts: []string{"9000/tcp"}, WaitingFor: wait.ForHTTP(
				"/minio/health/ready").WithPort("9000/tcp").WithStartupTimeout(2 * time.Minute),
			HostConfigModifier: func(config *container.HostConfig) {
				if config.PortBindings == nil {
					config.PortBindings = mobynetwork.PortMap{}
				}
				port := mobynetwork.MustParsePort("9000/tcp")
				config.PortBindings[port] = []mobynetwork.PortBinding{{
					HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: hostPort}}
			}}, Started: true})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		if err := instance.Terminate(cleanup); err != nil {
			t.Errorf("terminate MinIO: %v", err)
		}
	})
	host, err := instance.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := instance.MappedPort(ctx, "9000/tcp")
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"harpy-integration", "integration-secret-123", "")))
	if err != nil {
		t.Fatal(err)
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())
	client := s3.NewFromConfig(loaded, func(options *s3.Options) {
		options.BaseEndpoint, options.UsePathStyle = aws.String(endpoint), true
	})
	return client, endpoint
}
