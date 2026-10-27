//go:build integration

package s3blob

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"testing"
	"time"

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

const documentsMinIOImage = "minio/minio@sha256:" +
	"14cea493d9a34af32f524e538b8346cf79f3321eff8e708c1e2960462bd8936e"

func startDocumentsMinIO(t *testing.T) *s3.Client {
	t.Helper()
	hostPort := reserveDocumentsPort(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	runtime, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: documentsMinIOImage, Cmd: []string{"server", "/data", "--address", ":9000"},
			Env: map[string]string{"MINIO_ROOT_USER": "harpy-integration",
				"MINIO_ROOT_PASSWORD": "integration-secret-123"},
			ExposedPorts: []string{"9000/tcp"},
			HostConfigModifier: func(config *container.HostConfig) {
				if config.PortBindings == nil {
					config.PortBindings = mobynetwork.PortMap{}
				}
				port := mobynetwork.MustParsePort("9000/tcp")
				config.PortBindings[port] = []mobynetwork.PortBinding{{
					HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: hostPort}}
			},
			WaitingFor: wait.ForHTTP("/minio/health/ready").WithPort("9000/tcp").
				WithStartupTimeout(2 * time.Minute),
		}, Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		if err := runtime.Terminate(cleanup); err != nil {
			t.Errorf("terminate Documents MinIO: %v", err)
		}
	})
	host, err := runtime.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := runtime.MappedPort(ctx, "9000/tcp")
	if err != nil {
		t.Fatal(err)
	}
	endpoint := fmt.Sprintf("http://%s:%s", host, port.Port())
	loaded, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"harpy-integration", "integration-secret-123", "")))
	if err != nil {
		t.Fatal(err)
	}
	return s3.NewFromConfig(loaded, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})
}

func createDocumentsBucket(t *testing.T, client *s3.Client, bucket string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: &bucket}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.PutBucketVersioning(ctx, &s3.PutBucketVersioningInput{
		Bucket: &bucket, VersioningConfiguration: &types.VersioningConfiguration{
			Status: types.BucketVersioningStatusEnabled},
	}); err != nil {
		t.Fatal(err)
	}
}

func reserveDocumentsPort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return strconv.Itoa(port)
}
