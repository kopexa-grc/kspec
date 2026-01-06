package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"

	"github.com/kopexa-grc/kspec/core"
)

// LambdaFunctionResource fetches Lambda functions.
type LambdaFunctionResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *LambdaFunctionResource) Name() string {
	return "aws_lambda_function"
}

// Fetch retrieves all Lambda functions across configured regions.
func (r *LambdaFunctionResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := lambda.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := lambda.NewListFunctionsPaginator(client, &lambda.ListFunctionsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, fn := range page.Functions {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(fn.FunctionName)
				resource["arn"] = aws.ToString(fn.FunctionArn)
				resource["name"] = aws.ToString(fn.FunctionName)
				resource["region"] = region

				resource["runtime"] = string(fn.Runtime)
				resource["handler"] = aws.ToString(fn.Handler)
				resource["description"] = aws.ToString(fn.Description)
				resource["code_size"] = fn.CodeSize
				resource["timeout"] = fn.Timeout
				resource["memory_size"] = fn.MemorySize
				resource["last_modified"] = aws.ToString(fn.LastModified)

				// Role
				resource["role"] = aws.ToString(fn.Role)

				// VPC config
				if fn.VpcConfig != nil {
					resource["vpc_id"] = aws.ToString(fn.VpcConfig.VpcId)
					resource["subnet_ids"] = fn.VpcConfig.SubnetIds
					resource["security_group_ids"] = fn.VpcConfig.SecurityGroupIds
					resource["in_vpc"] = aws.ToString(fn.VpcConfig.VpcId) != ""
				} else {
					resource["in_vpc"] = false
				}

				// Environment
				if fn.Environment != nil && fn.Environment.Variables != nil {
					resource["environment_variables"] = fn.Environment.Variables
					resource["has_environment_variables"] = len(fn.Environment.Variables) > 0
				} else {
					resource["has_environment_variables"] = false
				}

				// KMS
				resource["kms_key_arn"] = aws.ToString(fn.KMSKeyArn)
				resource["has_kms_encryption"] = aws.ToString(fn.KMSKeyArn) != ""

				// Dead letter config
				if fn.DeadLetterConfig != nil {
					resource["dead_letter_target"] = aws.ToString(fn.DeadLetterConfig.TargetArn)
					resource["has_dead_letter_config"] = aws.ToString(fn.DeadLetterConfig.TargetArn) != ""
				} else {
					resource["has_dead_letter_config"] = false
				}

				// Tracing
				if fn.TracingConfig != nil {
					resource["tracing_mode"] = string(fn.TracingConfig.Mode)
					resource["has_xray_tracing"] = fn.TracingConfig.Mode == "Active"
				} else {
					resource["has_xray_tracing"] = false
				}

				// Layers
				var layerArns []string
				for _, layer := range fn.Layers {
					layerArns = append(layerArns, aws.ToString(layer.Arn))
				}
				resource["layers"] = layerArns
				resource["layer_count"] = len(layerArns)

				// Architectures
				var archs []string
				for _, arch := range fn.Architectures {
					archs = append(archs, string(arch))
				}
				resource["architectures"] = archs

				// Package type
				resource["package_type"] = string(fn.PackageType)
				resource["is_container"] = fn.PackageType == "Image"

				// Get function policy for public access check
				policyResp, err := client.GetPolicy(ctx, &lambda.GetPolicyInput{
					FunctionName: fn.FunctionName,
				})
				if err == nil && policyResp.Policy != nil {
					resource["has_resource_policy"] = true
					resource["resource_policy"] = aws.ToString(policyResp.Policy)
				} else {
					resource["has_resource_policy"] = false
				}

				// Get function URL config
				urlResp, err := client.GetFunctionUrlConfig(ctx, &lambda.GetFunctionUrlConfigInput{
					FunctionName: fn.FunctionName,
				})
				if err == nil {
					resource["function_url"] = aws.ToString(urlResp.FunctionUrl)
					resource["has_function_url"] = true
					resource["function_url_auth_type"] = string(urlResp.AuthType)
					resource["function_url_is_public"] = urlResp.AuthType == "NONE"
				} else {
					resource["has_function_url"] = false
				}

				// Computed: Check for deprecated runtimes
				runtime := string(fn.Runtime)
				deprecatedRuntimes := map[string]bool{
					"python2.7": true, "python3.6": true,
					"nodejs10.x": true, "nodejs12.x": true,
					"ruby2.5": true, "dotnetcore2.1": true,
				}
				resource["has_deprecated_runtime"] = deprecatedRuntimes[runtime]

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
