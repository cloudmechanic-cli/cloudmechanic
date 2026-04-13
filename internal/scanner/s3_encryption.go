package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3EncryptionAPI is the interface for S3 encryption checks.
type S3EncryptionAPI interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
	GetBucketEncryption(ctx context.Context, params *s3.GetBucketEncryptionInput, optFns ...func(*s3.Options)) (*s3.GetBucketEncryptionOutput, error)
	GetBucketVersioning(ctx context.Context, params *s3.GetBucketVersioningInput, optFns ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
}

// S3EncryptionScanner finds S3 buckets without default server-side encryption.
type S3EncryptionScanner struct {
	Client S3EncryptionAPI
}

func (s *S3EncryptionScanner) Name() string {
	return "S3 Buckets Without Encryption"
}

func (s *S3EncryptionScanner) Scan(ctx context.Context) ([]Issue, error) {
	listOut, err := s.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("s3 encryption scanner: %w", err)
	}

	var issues []Issue
	for _, bucket := range listOut.Buckets {
		name := aws.ToString(bucket.Name)

		_, err := s.Client.GetBucketEncryption(ctx, &s3.GetBucketEncryptionInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			if strings.Contains(err.Error(), "ServerSideEncryptionConfigurationNotFoundError") {
				issues = append(issues, Issue{
					Severity:    SeverityCritical,
					Scanner:     s.Name(),
					ResourceID:  name,
					Description: fmt.Sprintf("S3 bucket %s has no default server-side encryption configured", name),
					Suggestion:  "Enable default encryption with SSE-S3 (AES-256) or SSE-KMS on this bucket.",
				})
			}
			continue
		}
	}

	return issues, nil
}

// S3VersioningScanner finds S3 buckets without versioning enabled.
type S3VersioningScanner struct {
	Client S3EncryptionAPI
}

func (s *S3VersioningScanner) Name() string {
	return "S3 Buckets Without Versioning"
}

func (s *S3VersioningScanner) Scan(ctx context.Context) ([]Issue, error) {
	listOut, err := s.Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("s3 versioning scanner: %w", err)
	}

	var issues []Issue
	for _, bucket := range listOut.Buckets {
		name := aws.ToString(bucket.Name)

		verOut, err := s.Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{
			Bucket: bucket.Name,
		})
		if err != nil {
			continue
		}

		if verOut.Status != "Enabled" {
			issues = append(issues, Issue{
				Severity:    SeverityWarning,
				Scanner:     s.Name(),
				ResourceID:  name,
				Description: fmt.Sprintf("S3 bucket %s does not have versioning enabled", name),
				Suggestion:  "Enable versioning to protect against accidental deletes and overwrites.",
			})
		}
	}

	return issues, nil
}
