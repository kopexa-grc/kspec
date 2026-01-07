// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"go.uber.org/mock/gomock"

	"github.com/kopexa-grc/kspec/core"
)

func TestLambdaFunctionResource_Name(t *testing.T) {
	r := &LambdaFunctionResource{}
	if got := r.Name(); got != "aws_lambda_function" {
		t.Errorf("Name() = %v, want aws_lambda_function", got)
	}
}

func TestLambdaFunctionResource_Fetch_EmptyFunctions(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockLambdaAPI(ctrl)

	mock.EXPECT().
		ListFunctions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&lambda.ListFunctionsOutput{Functions: []types.FunctionConfiguration{}}, nil)

	r := &LambdaFunctionResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(resources))
	}
}

func TestLambdaFunctionResource_Fetch_SecureFunction(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockLambdaAPI(ctrl)

	fnName := "secure-function"
	fnArn := "arn:aws:lambda:us-east-1:123456789012:function:secure-function"

	mock.EXPECT().
		ListFunctions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&lambda.ListFunctionsOutput{
			Functions: []types.FunctionConfiguration{
				{
					FunctionName: aws.String(fnName),
					FunctionArn:  aws.String(fnArn),
					Runtime:      types.RuntimePython312,
					Handler:      aws.String("main.handler"),
					Description:  aws.String("A secure Lambda function"),
					CodeSize:     1024000,
					Timeout:      aws.Int32(30),
					MemorySize:   aws.Int32(256),
					Role:         aws.String("arn:aws:iam::123456789012:role/lambda-role"),
					VpcConfig: &types.VpcConfigResponse{
						VpcId:            aws.String("vpc-12345"),
						SubnetIds:        []string{"subnet-1", "subnet-2"},
						SecurityGroupIds: []string{"sg-lambda"},
					},
					KMSKeyArn: aws.String("arn:aws:kms:us-east-1:123456789012:key/12345"),
					DeadLetterConfig: &types.DeadLetterConfig{
						TargetArn: aws.String("arn:aws:sqs:us-east-1:123456789012:dlq"),
					},
					TracingConfig: &types.TracingConfigResponse{
						Mode: types.TracingModeActive,
					},
					PackageType:   types.PackageTypeZip,
					Architectures: []types.Architecture{types.ArchitectureArm64},
				},
			},
		}, nil)

	// No resource policy
	mock.EXPECT().
		GetPolicy(gomock.Any(), gomock.Any()).
		Return(nil, &types.ResourceNotFoundException{})

	// No function URL
	mock.EXPECT().
		GetFunctionUrlConfig(gomock.Any(), gomock.Any()).
		Return(nil, &types.ResourceNotFoundException{})

	r := &LambdaFunctionResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]
	if res["name"] != fnName {
		t.Errorf("name = %v, want %v", res["name"], fnName)
	}
	if res["in_vpc"] != true {
		t.Errorf("in_vpc = %v, want true", res["in_vpc"])
	}
	if res["has_kms_encryption"] != true {
		t.Errorf("has_kms_encryption = %v, want true", res["has_kms_encryption"])
	}
	if res["has_dead_letter_config"] != true {
		t.Errorf("has_dead_letter_config = %v, want true", res["has_dead_letter_config"])
	}
	if res["has_xray_tracing"] != true {
		t.Errorf("has_xray_tracing = %v, want true", res["has_xray_tracing"])
	}
	if res["has_deprecated_runtime"] != false {
		t.Errorf("has_deprecated_runtime = %v, want false", res["has_deprecated_runtime"])
	}
}

func TestLambdaFunctionResource_Fetch_DeprecatedRuntime(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockLambdaAPI(ctrl)

	mock.EXPECT().
		ListFunctions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&lambda.ListFunctionsOutput{
			Functions: []types.FunctionConfiguration{
				{
					FunctionName: aws.String("legacy-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:legacy-function"),
					Runtime:      types.RuntimePython27, // Deprecated!
					Handler:      aws.String("main.handler"),
					PackageType:  types.PackageTypeZip,
				},
			},
		}, nil)

	mock.EXPECT().GetPolicy(gomock.Any(), gomock.Any()).Return(nil, &types.ResourceNotFoundException{})
	mock.EXPECT().GetFunctionUrlConfig(gomock.Any(), gomock.Any()).Return(nil, &types.ResourceNotFoundException{})

	r := &LambdaFunctionResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["has_deprecated_runtime"] != true {
		t.Errorf("has_deprecated_runtime = %v, want true for python2.7", res["has_deprecated_runtime"])
	}
}

func TestLambdaFunctionResource_Fetch_PublicFunctionURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockLambdaAPI(ctrl)

	fnName := "public-url-function"

	mock.EXPECT().
		ListFunctions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&lambda.ListFunctionsOutput{
			Functions: []types.FunctionConfiguration{
				{
					FunctionName: aws.String(fnName),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:public-url-function"),
					Runtime:      types.RuntimeNodejs20x,
					Handler:      aws.String("index.handler"),
					PackageType:  types.PackageTypeZip,
				},
			},
		}, nil)

	mock.EXPECT().GetPolicy(gomock.Any(), gomock.Any()).Return(nil, &types.ResourceNotFoundException{})

	// Has public function URL
	mock.EXPECT().
		GetFunctionUrlConfig(gomock.Any(), gomock.Any()).
		Return(&lambda.GetFunctionUrlConfigOutput{
			FunctionUrl: aws.String("https://abc123.lambda-url.us-east-1.on.aws/"),
			AuthType:    types.FunctionUrlAuthTypeNone, // Public!
		}, nil)

	r := &LambdaFunctionResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["has_function_url"] != true {
		t.Errorf("has_function_url = %v, want true", res["has_function_url"])
	}
	if res["function_url_is_public"] != true {
		t.Errorf("function_url_is_public = %v, want true", res["function_url_is_public"])
	}
}

func TestLambdaFunctionResource_Fetch_ContainerImage(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockLambdaAPI(ctrl)

	mock.EXPECT().
		ListFunctions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&lambda.ListFunctionsOutput{
			Functions: []types.FunctionConfiguration{
				{
					FunctionName: aws.String("container-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:container-function"),
					PackageType:  types.PackageTypeImage,
				},
			},
		}, nil)

	mock.EXPECT().GetPolicy(gomock.Any(), gomock.Any()).Return(nil, &types.ResourceNotFoundException{})
	mock.EXPECT().GetFunctionUrlConfig(gomock.Any(), gomock.Any()).Return(nil, &types.ResourceNotFoundException{})

	r := &LambdaFunctionResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["is_container"] != true {
		t.Errorf("is_container = %v, want true", res["is_container"])
	}
}

func TestLambdaFunctionResource_Fetch_WithResourcePolicy(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockLambdaAPI(ctrl)

	policyDoc := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":"*","Action":"lambda:InvokeFunction","Resource":"*"}]}`

	mock.EXPECT().
		ListFunctions(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&lambda.ListFunctionsOutput{
			Functions: []types.FunctionConfiguration{
				{
					FunctionName: aws.String("policy-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:policy-function"),
					Runtime:      types.RuntimeNodejs20x,
					PackageType:  types.PackageTypeZip,
				},
			},
		}, nil)

	mock.EXPECT().
		GetPolicy(gomock.Any(), gomock.Any()).
		Return(&lambda.GetPolicyOutput{
			Policy: aws.String(policyDoc),
		}, nil)

	mock.EXPECT().GetFunctionUrlConfig(gomock.Any(), gomock.Any()).Return(nil, &types.ResourceNotFoundException{})

	r := &LambdaFunctionResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["has_resource_policy"] != true {
		t.Errorf("has_resource_policy = %v, want true", res["has_resource_policy"])
	}
	if res["resource_policy"] != policyDoc {
		t.Errorf("resource_policy doesn't match expected policy document")
	}
}
