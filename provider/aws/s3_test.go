package aws

import (
	"errors"
	"strings"
	"testing"

	"github.com/kopexa-grc/kspec/core"
)

func TestS3BucketResource_Name(t *testing.T) {
	r := &S3BucketResource{}
	if got := r.Name(); got != "aws_s3_bucket" {
		t.Errorf("Name() = %v, want aws_s3_bucket", got)
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

// TestProviderErrorWrapping tests that provider errors include context
func TestProviderErrorWrapping(t *testing.T) {
	originalErr := errors.New("access denied")
	wrappedErr := core.WrapError("aws", "s3_bucket", "list", originalErr)

	// Check that the error message includes context
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

	// Check that original error is preserved
	if !errors.Is(wrappedErr, originalErr) {
		t.Error("wrapped error should preserve original error")
	}
}
