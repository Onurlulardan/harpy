package s3object

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const sdkRetryAttempts = 3

type sdkClientSettings struct {
	endpoint, region                string
	pathStyle, static               bool
	accessKeyID, secretKey, session string
}

func NewWorkloadIdentityClient(ctx context.Context, endpoint, region string,
	pathStyle bool) (*s3.Client, error) {
	return newSDKClient(ctx, sdkClientSettings{endpoint: endpoint,
		region: region, pathStyle: pathStyle})
}

func NewStaticCredentialsClient(ctx context.Context, endpoint, region string,
	pathStyle bool, accessKeyID, secretKey, sessionToken string) (*s3.Client, error) {
	return newSDKClient(ctx, sdkClientSettings{endpoint: endpoint, region: region,
		pathStyle: pathStyle, static: true, accessKeyID: accessKeyID,
		secretKey: secretKey, session: sessionToken})
}

func newSDKClient(ctx context.Context, settings sdkClientSettings) (*s3.Client, error) {
	if ctx == nil || ctx.Err() != nil || !validSDKEndpoint(settings.endpoint) ||
		!validSDKRegion(settings.region) || settings.static &&
		(!validSDKCredential(settings.accessKeyID) || !validSDKCredential(settings.secretKey) ||
			settings.session != "" && !validSDKCredential(settings.session)) {
		return nil, ErrConfiguration
	}
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(settings.region),
		awsconfig.WithRetryMode(aws.RetryModeStandard),
		awsconfig.WithRetryMaxAttempts(sdkRetryAttempts),
	}
	if settings.static {
		provider := credentials.NewStaticCredentialsProvider(settings.accessKeyID,
			settings.secretKey, settings.session)
		options = append(options, awsconfig.WithCredentialsProvider(provider))
	}
	loaded, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("load artifact S3 SDK config: %w", err)
	}
	if !settings.static {
		loaded.Credentials = aws.NewCredentialsCache(workloadIdentityProvider{
			provider: loaded.Credentials})
	}
	return s3.NewFromConfig(loaded, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(settings.endpoint)
		options.UsePathStyle = settings.pathStyle
		options.DisableMultiRegionAccessPoints = true
	}), nil
}

func validSDKEndpoint(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed != nil && parsed.Host != "" && parsed.User == nil &&
		(parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Path == "" &&
		parsed.RawQuery == "" && parsed.Fragment == "" && parsed.String() == value
}

func validSDKRegion(value string) bool {
	if len(value) < 1 || len(value) > 63 || value != strings.ToLower(value) {
		return false
	}
	for index := range len(value) {
		if (value[index] < 'a' || value[index] > 'z') &&
			(value[index] < '0' || value[index] > '9') && value[index] != '-' {
			return false
		}
	}
	return value[0] != '-'
}

func validSDKCredential(value string) bool {
	if value == "" {
		return false
	}
	for index := range len(value) {
		if value[index] < 0x21 || value[index] > 0x7e {
			return false
		}
	}
	return true
}
