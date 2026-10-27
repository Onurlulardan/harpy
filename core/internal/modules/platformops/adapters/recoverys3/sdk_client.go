package recoverys3

import (
	"context"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const sdkRetryAttempts = 3

func NewStaticClient(
	ctx context.Context,
	endpoint string,
	region string,
	pathStyle bool,
	tlsCASHA256 string,
	credential Credential,
) (*s3.Client, error) {
	parsed, err := url.Parse(endpoint)
	if ctx == nil || ctx.Err() != nil || err != nil ||
		parsed == nil || parsed.User != nil || parsed.Host == "" ||
		(parsed.Scheme != "https" && parsed.Scheme != "http") ||
		parsed.Path != "" || parsed.RawQuery != "" ||
		parsed.Fragment != "" || parsed.String() != endpoint ||
		!validRegion(region) ||
		credential.Schema != CredentialSchema ||
		!validCredentialValue(credential.AccessKeyID) ||
		!validCredentialValue(credential.SecretAccessKey) ||
		credential.SessionToken != "" &&
			!validCredentialValue(credential.SessionToken) {
		return nil, ErrObject
	}
	httpClient, err := pinnedHTTPClient(parsed.Scheme, tlsCASHA256)
	if err != nil {
		return nil, err
	}
	provider := credentials.NewStaticCredentialsProvider(
		credential.AccessKeyID,
		credential.SecretAccessKey,
		credential.SessionToken,
	)
	config, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithRetryMode(aws.RetryModeStandard),
		awsconfig.WithRetryMaxAttempts(sdkRetryAttempts),
		awsconfig.WithCredentialsProvider(provider),
		awsconfig.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("load recovery S3 config: %w", err)
	}
	return s3.NewFromConfig(config, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = pathStyle
		options.DisableMultiRegionAccessPoints = true
	}), nil
}

func validRegion(value string) bool {
	if len(value) < 1 || len(value) > 63 {
		return false
	}
	for index, character := range value {
		if (character < 'a' || character > 'z') &&
			(character < '0' || character > '9') && character != '-' ||
			index == 0 && character == '-' {
			return false
		}
	}
	return value[len(value)-1] != '-'
}
