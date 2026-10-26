package main

import (
	"errors"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type answers struct {
	Profile              string
	HTTPPort             int
	PostgresPort         int
	PostgresDB           string
	PostgresUser         string
	PostgresPassword     string
	RabbitPort           int
	RabbitManagementPort int
	RabbitUser           string
	RabbitPassword       string
	S3Endpoint           string
	S3Region             string
	S3Bucket             string
	S3AccessKey          string
	S3SecretKey          string
	S3PathStyle          bool
	S3Provider           string
	RustFSAccessKey      string
	RustFSSecretKey      string
}

func defaultAnswers() answers {
	return answers{
		Profile:              "single-tenant",
		HTTPPort:             8080,
		PostgresPort:         5432,
		PostgresDB:           "harpy",
		PostgresUser:         "harpy",
		RabbitPort:           5672,
		RabbitManagementPort: 15672,
		RabbitUser:           "harpy",
		S3Endpoint:           "http://127.0.0.1:9000",
		S3Region:             "us-east-1",
		S3Bucket:             "harpy",
		S3PathStyle:          true,
		S3Provider:           "rustfs",
	}
}

func tenancyMode(profile string) (string, error) {
	switch profile {
	case "single-tenant":
		return "single", nil
	case "multi-tenant":
		return "multi", nil
	default:
		return "", errors.New("profile must be single-tenant or multi-tenant")
	}
}

func validProfile(value string) error {
	_, err := tenancyMode(value)
	return err
}

func validPort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return errors.New("port must be between 1 and 65535")
	}
	return nil
}

func validIdent(kind, value string) error {
	if !identPattern.MatchString(value) {
		return errors.New(kind + " must be a letter or underscore followed by letters, digits, or underscores")
	}
	return nil
}

func validEndpoint(raw string) error {
	parsed, err := url.ParseRequestURI(raw)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return errors.New("endpoint must be an http or https URL")
	}
	return nil
}

func validRegion(value string) error {
	if !regionPattern.MatchString(value) {
		return errors.New("region must use lowercase letters, digits, and hyphens")
	}
	return nil
}

func validBucket(value string) error {
	if !bucketPattern.MatchString(value) {
		return errors.New("bucket must be a lowercase DNS name")
	}
	return nil
}

func validChoice(value string, allowed ...string) error {
	for _, item := range allowed {
		if value == item {
			return nil
		}
	}
	return errors.New("choose " + strings.Join(allowed, " or "))
}

func (a answers) normalized() answers {
	switch a.S3Provider {
	case "rustfs":
		a.S3PathStyle = true
		a.S3Endpoint = "http://127.0.0.1:9000"
		a.S3AccessKey = a.RustFSAccessKey
		a.S3SecretKey = a.RustFSSecretKey
	case "minio":
		a.S3PathStyle = true
	case "aws":
		a.S3PathStyle = false
	}
	return a
}

func (a answers) validate() error {
	if err := validProfile(a.Profile); err != nil {
		return err
	}
	if err := validPort(strconv.Itoa(a.HTTPPort)); err != nil {
		return err
	}
	if err := validPort(strconv.Itoa(a.PostgresPort)); err != nil {
		return err
	}
	if err := validPort(strconv.Itoa(a.RabbitPort)); err != nil || validPort(strconv.Itoa(a.RabbitManagementPort)) != nil {
		return errors.New("rabbitmq ports are invalid")
	}
	if err := validIdent("database", a.PostgresDB); err != nil {
		return err
	}
	if err := validIdent("user", a.PostgresUser); err != nil {
		return err
	}
	if err := validIdent("user", a.RabbitUser); err != nil {
		return err
	}
	if a.PostgresPassword == "" || a.RabbitPassword == "" || a.RustFSSecretKey == "" {
		return errors.New("generated secrets are missing")
	}
	if err := validateAccessKey(a.RustFSAccessKey); err != nil {
		return err
	}
	if strings.EqualFold(a.RustFSSecretKey, "rustfsadmin") || strings.EqualFold(a.PostgresPassword, "rustfsadmin") {
		return errors.New("refusing rustfsadmin credential")
	}
	switch a.S3Provider {
	case "rustfs", "aws", "minio":
	default:
		return errors.New("S3 provider must be rustfs, aws, or minio")
	}
	if err := validEndpoint(a.S3Endpoint); err != nil {
		return err
	}
	if err := validRegion(a.S3Region); err != nil {
		return err
	}
	if err := validBucket(a.S3Bucket); err != nil {
		return err
	}
	if a.S3AccessKey == "" || a.S3SecretKey == "" {
		return errors.New("S3 credentials are required")
	}
	if strings.EqualFold(a.S3AccessKey, "rustfsadmin") || strings.EqualFold(a.S3SecretKey, "rustfsadmin") {
		return errors.New("refusing rustfsadmin credential")
	}
	if a.S3Provider != "aws" && !a.S3PathStyle {
		return errors.New("path-style must be on for rustfs and minio")
	}
	if a.S3Provider == "aws" && a.S3PathStyle {
		return errors.New("path-style must be off for aws")
	}
	if a.S3Provider == "rustfs" {
		return validateAccessKey(a.S3AccessKey)
	}
	return nil
}

var (
	identPattern  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,62}$`)
	regionPattern = regexp.MustCompile(`^[a-z0-9-]+$`)
	bucketPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`)
)
