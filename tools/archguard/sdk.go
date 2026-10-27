package main

import "strings"

// Blob code is an object-storage adapter. It may import aws-sdk-go-v2.
// MinIO and RustFS SDK modules are rejected in every file.
func objectSDKViolation(importPath string) string {
	if modulePrefix(importPath, "github.com/minio/minio-go") ||
		modulePrefix(importPath, "github.com/minio/madmin-go") ||
		modulePrefix(importPath, "github.com/minio/minio") {
		return "MinIO SDK imports are forbidden"
	}
	if rustfsSDK(importPath) {
		return "RustFS SDK imports are forbidden"
	}
	return ""
}

func awsGate(source guardedSource, importPath string) string {
	if modulePrefix(importPath, "github.com/aws/aws-sdk-go-v2") && !blobCode(source.path) {
		return "aws-sdk-go-v2 imports are limited to blob code"
	}
	return ""
}

func domainViolation(importPath string) string {
	switch {
	case importPath == "net/http" || strings.HasPrefix(importPath, "net/http/"):
		return "domain cannot import net/http"
	case importPath == "database/sql" || strings.HasPrefix(importPath, "database/sql/"):
		return "domain cannot import database/sql"
	case modulePrefix(importPath, "github.com/jackc/pgx"):
		return "domain cannot import pgx"
	case cloudSDK(importPath):
		return "domain cannot import cloud SDKs"
	default:
		return ""
	}
}

func cloudSDK(importPath string) bool {
	if rustfsSDK(importPath) {
		return true
	}
	for _, module := range []string{
		"github.com/aws/aws-sdk-go-v2",
		"github.com/aws/aws-sdk-go",
		"github.com/aws/smithy-go",
		"cloud.google.com",
		"google.golang.org/api",
		"github.com/Azure/azure-sdk-for-go",
		"github.com/minio/minio-go",
		"github.com/minio/madmin-go",
		"github.com/minio/minio",
	} {
		if modulePrefix(importPath, module) {
			return true
		}
	}
	return false
}

func blobCode(sourcePath string) bool {
	for _, part := range strings.Split(strings.ToLower(sourcePath), "/") {
		if strings.Contains(part, "blob") || strings.Contains(part, "s3") ||
			strings.Contains(part, "artifact") {
			return true
		}
	}
	return false
}

func modulePrefix(importPath, module string) bool {
	return importPath == module || strings.HasPrefix(importPath, module+"/")
}

func rustfsSDK(importPath string) bool {
	if importPath == modulePath || strings.HasPrefix(importPath, modulePath+"/") {
		return false
	}
	for _, part := range strings.Split(strings.ToLower(importPath), "/") {
		if part == "rustfs" || strings.HasPrefix(part, "rustfs-") || strings.HasSuffix(part, "-rustfs") {
			return true
		}
	}
	return false
}
