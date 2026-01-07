// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package aws

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/mock/gomock"

	"github.com/kopexa-grc/kspec/core"
)

func TestS3BucketResource_Name(t *testing.T) {
	r := &S3BucketResource{}
	if got := r.Name(); got != "aws_s3_bucket" {
		t.Errorf("Name() = %v, want aws_s3_bucket", got)
	}
}

func TestS3BucketResource_Fetch_EmptyBuckets(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockS3API(ctrl)
	mock.EXPECT().
		ListBuckets(gomock.Any(), gomock.Any()).
		Return(&s3.ListBucketsOutput{Buckets: []types.Bucket{}}, nil)

	r := &S3BucketResource{client: mock}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(resources))
	}
}

func TestS3BucketResource_Fetch_SingleBucket(t *testing.T) {
	ctrl := gomock.NewController(t)

	createdAt := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)

	mock := NewMockS3API(ctrl)

	// ListBuckets returns one bucket
	mock.EXPECT().
		ListBuckets(gomock.Any(), gomock.Any()).
		Return(&s3.ListBucketsOutput{
			Buckets: []types.Bucket{
				{Name: aws.String("test-bucket"), CreationDate: &createdAt},
			},
		}, nil)

	// Expect all the bucket detail calls
	mock.EXPECT().
		GetBucketLocation(gomock.Any(), gomock.Any()).
		Return(&s3.GetBucketLocationOutput{
			LocationConstraint: types.BucketLocationConstraintEuWest1,
		}, nil)

	mock.EXPECT().
		GetBucketVersioning(gomock.Any(), gomock.Any()).
		Return(&s3.GetBucketVersioningOutput{
			Status: types.BucketVersioningStatusEnabled,
		}, nil)

	mock.EXPECT().
		GetBucketEncryption(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no encryption"))

	mock.EXPECT().
		GetBucketLogging(gomock.Any(), gomock.Any()).
		Return(&s3.GetBucketLoggingOutput{}, nil)

	mock.EXPECT().
		GetPublicAccessBlock(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no public access block"))

	mock.EXPECT().
		GetBucketAcl(gomock.Any(), gomock.Any()).
		Return(&s3.GetBucketAclOutput{}, nil)

	mock.EXPECT().
		GetBucketPolicy(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no policy"))

	mock.EXPECT().
		GetBucketLifecycleConfiguration(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no lifecycle"))

	mock.EXPECT().
		GetObjectLockConfiguration(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("no object lock"))

	r := &S3BucketResource{client: mock}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]

	// Check basic properties
	if got := res["id"]; got != "test-bucket" {
		t.Errorf("id = %v, want %q", got, "test-bucket")
	}
	if got := res["arn"]; got != "arn:aws:s3:::test-bucket" {
		t.Errorf("arn = %v, want %q", got, "arn:aws:s3:::test-bucket")
	}
	if got := res["region"]; got != "eu-west-1" {
		t.Errorf("region = %v, want %q", got, "eu-west-1")
	}
	if got := res["has_versioning"]; got != true {
		t.Errorf("has_versioning = %v, want true", got)
	}
}

func TestS3BucketResource_Fetch_WithEncryption(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockS3API(ctrl)

	mock.EXPECT().
		ListBuckets(gomock.Any(), gomock.Any()).
		Return(&s3.ListBucketsOutput{
			Buckets: []types.Bucket{{Name: aws.String("encrypted-bucket")}},
		}, nil)

	mock.EXPECT().GetBucketLocation(gomock.Any(), gomock.Any()).Return(&s3.GetBucketLocationOutput{}, nil)
	mock.EXPECT().GetBucketVersioning(gomock.Any(), gomock.Any()).Return(&s3.GetBucketVersioningOutput{}, nil)

	mock.EXPECT().
		GetBucketEncryption(gomock.Any(), gomock.Any()).
		Return(&s3.GetBucketEncryptionOutput{
			ServerSideEncryptionConfiguration: &types.ServerSideEncryptionConfiguration{
				Rules: []types.ServerSideEncryptionRule{
					{
						ApplyServerSideEncryptionByDefault: &types.ServerSideEncryptionByDefault{
							SSEAlgorithm:   types.ServerSideEncryptionAwsKms,
							KMSMasterKeyID: aws.String("arn:aws:kms:us-east-1:123456789012:key/test-key"),
						},
						BucketKeyEnabled: aws.Bool(true),
					},
				},
			},
		}, nil)

	mock.EXPECT().GetBucketLogging(gomock.Any(), gomock.Any()).Return(&s3.GetBucketLoggingOutput{}, nil)
	mock.EXPECT().GetPublicAccessBlock(gomock.Any(), gomock.Any()).Return(nil, errors.New("no pab"))
	mock.EXPECT().GetBucketAcl(gomock.Any(), gomock.Any()).Return(&s3.GetBucketAclOutput{}, nil)
	mock.EXPECT().GetBucketPolicy(gomock.Any(), gomock.Any()).Return(nil, errors.New("no policy"))
	mock.EXPECT().GetBucketLifecycleConfiguration(gomock.Any(), gomock.Any()).Return(nil, errors.New("no lifecycle"))
	mock.EXPECT().GetObjectLockConfiguration(gomock.Any(), gomock.Any()).Return(nil, errors.New("no object lock"))

	r := &S3BucketResource{client: mock}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]

	if got := res["has_encryption"]; got != true {
		t.Errorf("has_encryption = %v, want true", got)
	}
	if got := res["has_kms_encryption"]; got != true {
		t.Errorf("has_kms_encryption = %v, want true", got)
	}
	if got := res["bucket_key_enabled"]; got != true {
		t.Errorf("bucket_key_enabled = %v, want true", got)
	}
}

func TestS3BucketResource_Fetch_WithPublicAccessBlock(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockS3API(ctrl)

	mock.EXPECT().
		ListBuckets(gomock.Any(), gomock.Any()).
		Return(&s3.ListBucketsOutput{
			Buckets: []types.Bucket{{Name: aws.String("secure-bucket")}},
		}, nil)

	mock.EXPECT().GetBucketLocation(gomock.Any(), gomock.Any()).Return(&s3.GetBucketLocationOutput{}, nil)
	mock.EXPECT().GetBucketVersioning(gomock.Any(), gomock.Any()).Return(&s3.GetBucketVersioningOutput{}, nil)
	mock.EXPECT().GetBucketEncryption(gomock.Any(), gomock.Any()).Return(nil, errors.New("no encryption"))
	mock.EXPECT().GetBucketLogging(gomock.Any(), gomock.Any()).Return(&s3.GetBucketLoggingOutput{}, nil)

	mock.EXPECT().
		GetPublicAccessBlock(gomock.Any(), gomock.Any()).
		Return(&s3.GetPublicAccessBlockOutput{
			PublicAccessBlockConfiguration: &types.PublicAccessBlockConfiguration{
				BlockPublicAcls:       aws.Bool(true),
				IgnorePublicAcls:      aws.Bool(true),
				BlockPublicPolicy:     aws.Bool(true),
				RestrictPublicBuckets: aws.Bool(true),
			},
		}, nil)

	mock.EXPECT().GetBucketAcl(gomock.Any(), gomock.Any()).Return(&s3.GetBucketAclOutput{}, nil)
	mock.EXPECT().GetBucketPolicy(gomock.Any(), gomock.Any()).Return(nil, errors.New("no policy"))
	mock.EXPECT().GetBucketLifecycleConfiguration(gomock.Any(), gomock.Any()).Return(nil, errors.New("no lifecycle"))
	mock.EXPECT().GetObjectLockConfiguration(gomock.Any(), gomock.Any()).Return(nil, errors.New("no object lock"))

	r := &S3BucketResource{client: mock}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]

	if got := res["block_public_acls"]; got != true {
		t.Errorf("block_public_acls = %v, want true", got)
	}
	if got := res["public_access_block_enabled"]; got != true {
		t.Errorf("public_access_block_enabled = %v, want true", got)
	}
}

func TestS3BucketResource_Fetch_ListBucketsError(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockS3API(ctrl)
	mock.EXPECT().
		ListBuckets(gomock.Any(), gomock.Any()).
		Return(nil, errors.New("access denied"))

	r := &S3BucketResource{client: mock}
	_, err := r.Fetch(context.Background(), core.Asset{})

	if err == nil {
		t.Error("Fetch() expected error, got nil")
	}
}

func TestS3BucketResource_Fetch_PublicBucket(t *testing.T) {
	ctrl := gomock.NewController(t)

	mock := NewMockS3API(ctrl)

	mock.EXPECT().
		ListBuckets(gomock.Any(), gomock.Any()).
		Return(&s3.ListBucketsOutput{
			Buckets: []types.Bucket{{Name: aws.String("public-bucket")}},
		}, nil)

	mock.EXPECT().GetBucketLocation(gomock.Any(), gomock.Any()).Return(&s3.GetBucketLocationOutput{}, nil)
	mock.EXPECT().GetBucketVersioning(gomock.Any(), gomock.Any()).Return(&s3.GetBucketVersioningOutput{}, nil)
	mock.EXPECT().GetBucketEncryption(gomock.Any(), gomock.Any()).Return(nil, errors.New("no encryption"))
	mock.EXPECT().GetBucketLogging(gomock.Any(), gomock.Any()).Return(&s3.GetBucketLoggingOutput{}, nil)
	mock.EXPECT().GetPublicAccessBlock(gomock.Any(), gomock.Any()).Return(nil, errors.New("no pab"))

	mock.EXPECT().
		GetBucketAcl(gomock.Any(), gomock.Any()).
		Return(&s3.GetBucketAclOutput{
			Grants: []types.Grant{
				{
					Grantee: &types.Grantee{
						URI: aws.String("http://acs.amazonaws.com/groups/global/AllUsers"),
					},
					Permission: types.PermissionRead,
				},
			},
		}, nil)

	mock.EXPECT().GetBucketPolicy(gomock.Any(), gomock.Any()).Return(nil, errors.New("no policy"))
	mock.EXPECT().GetBucketLifecycleConfiguration(gomock.Any(), gomock.Any()).Return(nil, errors.New("no lifecycle"))
	mock.EXPECT().GetObjectLockConfiguration(gomock.Any(), gomock.Any()).Return(nil, errors.New("no object lock"))

	r := &S3BucketResource{client: mock}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]

	if got := res["has_public_acl"]; got != true {
		t.Errorf("has_public_acl = %v, want true", got)
	}
	if got := res["is_public"]; got != true {
		t.Errorf("is_public = %v, want true", got)
	}
}

func TestCheckPublicBucketPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy map[string]any
		want   bool
	}{
		{
			name:   "empty policy",
			policy: map[string]any{},
			want:   false,
		},
		{
			name: "public principal without condition",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect":    "Allow",
						"Principal": "*",
					},
				},
			},
			want: true,
		},
		{
			name: "public principal with condition",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect":    "Allow",
						"Principal": "*",
						"Condition": map[string]any{
							"IpAddress": map[string]any{
								"aws:SourceIp": "192.168.1.0/24",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "deny statement with public principal",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect":    "Deny",
						"Principal": "*",
					},
				},
			},
			want: false,
		},
		{
			name: "specific principal",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect":    "Allow",
						"Principal": map[string]any{"AWS": "arn:aws:iam::123456789012:root"},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkPublicBucketPolicy(tt.policy); got != tt.want {
				t.Errorf("checkPublicBucketPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckSecureTransportPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy map[string]any
		want   bool
	}{
		{
			name:   "empty policy",
			policy: map[string]any{},
			want:   false,
		},
		{
			name: "requires secure transport",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect": "Deny",
						"Condition": map[string]any{
							"Bool": map[string]any{
								"aws:SecureTransport": "false",
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "no secure transport requirement",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect": "Allow",
						"Condition": map[string]any{
							"Bool": map[string]any{
								"aws:SecureTransport": "false",
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "secure transport true (not enforcing)",
			policy: map[string]any{
				"Statement": []any{
					map[string]any{
						"Effect": "Deny",
						"Condition": map[string]any{
							"Bool": map[string]any{
								"aws:SecureTransport": "true",
							},
						},
					},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkSecureTransportPolicy(tt.policy); got != tt.want {
				t.Errorf("checkSecureTransportPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProviderErrorWrapping(t *testing.T) {
	originalErr := errors.New("access denied")
	wrappedErr := core.WrapError("aws", "s3_bucket", "list", originalErr)

	errMsg := wrappedErr.Error()
	if !strings.Contains(errMsg, "aws") {
		t.Errorf("error should contain provider name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "s3_bucket") {
		t.Errorf("error should contain resource name, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "list") {
		t.Errorf("error should contain operation, got: %s", errMsg)
	}
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("wrapped error should preserve original error")
	}
}
