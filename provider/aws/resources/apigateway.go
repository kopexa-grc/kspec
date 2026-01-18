// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"

	"github.com/kopexa-grc/kspec/core"
)

// APIGatewayClient is the interface for API Gateway REST API operations.
type APIGatewayClient interface {
	GetRestApis(ctx context.Context, params *apigateway.GetRestApisInput, optFns ...func(*apigateway.Options)) (*apigateway.GetRestApisOutput, error)
	GetStages(ctx context.Context, params *apigateway.GetStagesInput, optFns ...func(*apigateway.Options)) (*apigateway.GetStagesOutput, error)
	GetResources(ctx context.Context, params *apigateway.GetResourcesInput, optFns ...func(*apigateway.Options)) (*apigateway.GetResourcesOutput, error)
}

// APIGatewayV2Client is the interface for API Gateway V2 (HTTP/WebSocket) operations.
//
//nolint:dupl // AWS SDK client interfaces have similar structures by design
type APIGatewayV2Client interface {
	GetApis(ctx context.Context, params *apigatewayv2.GetApisInput, optFns ...func(*apigatewayv2.Options)) (*apigatewayv2.GetApisOutput, error)
	GetStages(ctx context.Context, params *apigatewayv2.GetStagesInput, optFns ...func(*apigatewayv2.Options)) (*apigatewayv2.GetStagesOutput, error)
	GetRoutes(ctx context.Context, params *apigatewayv2.GetRoutesInput, optFns ...func(*apigatewayv2.Options)) (*apigatewayv2.GetRoutesOutput, error)
	GetIntegrations(ctx context.Context, params *apigatewayv2.GetIntegrationsInput, optFns ...func(*apigatewayv2.Options)) (*apigatewayv2.GetIntegrationsOutput, error)
	GetAuthorizers(ctx context.Context, params *apigatewayv2.GetAuthorizersInput, optFns ...func(*apigatewayv2.Options)) (*apigatewayv2.GetAuthorizersOutput, error)
}

// APIGatewayClientFactory creates API Gateway clients for specific regions.
type APIGatewayClientFactory func(region string) APIGatewayClient

// APIGatewayV2ClientFactory creates API Gateway V2 clients for specific regions.
type APIGatewayV2ClientFactory func(region string) APIGatewayV2Client

// APIGatewayRestAPI fetches API Gateway REST APIs.
type APIGatewayRestAPI struct {
	clientFactory APIGatewayClientFactory
	accountID     string
	regions       []string
}

// NewAPIGatewayRestAPI creates a new APIGatewayRestAPI resource.
func NewAPIGatewayRestAPI(clientFactory APIGatewayClientFactory, accountID string, regions []string) *APIGatewayRestAPI {
	return &APIGatewayRestAPI{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *APIGatewayRestAPI) Name() string {
	return "aws_apigateway_restapi"
}

// Fetch retrieves all API Gateway REST APIs across configured regions.
func (r *APIGatewayRestAPI) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := apigateway.NewGetRestApisPaginator(client, &apigateway.GetRestApisInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, api := range page.Items {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(api.Id)
				resource["name"] = aws.ToString(api.Name)
				resource["region"] = region

				resource["description"] = aws.ToString(api.Description)
				resource["api_key_source"] = string(api.ApiKeySource)
				resource["endpoint_configuration_types"] = api.EndpointConfiguration.Types
				resource["minimum_compression_size"] = api.MinimumCompressionSize
				resource["policy"] = aws.ToString(api.Policy)
				resource["version"] = aws.ToString(api.Version)

				if api.CreatedDate != nil {
					resource["created_date"] = api.CreatedDate.String()
				}

				// Check for VPC endpoint IDs
				if api.EndpointConfiguration != nil {
					resource["vpc_endpoint_ids"] = api.EndpointConfiguration.VpcEndpointIds
					resource["has_vpc_endpoints"] = len(api.EndpointConfiguration.VpcEndpointIds) > 0

					// Endpoint types
					types := make([]string, 0, len(api.EndpointConfiguration.Types))
					for _, t := range api.EndpointConfiguration.Types {
						types = append(types, string(t))
					}
					resource["endpoint_types"] = types
					resource["is_regional"] = containsEndpointType(types, "REGIONAL")
					resource["is_edge_optimized"] = containsEndpointType(types, "EDGE")
					resource["is_private"] = containsEndpointType(types, "PRIVATE")
				}

				// Disable execute API endpoint
				resource["disable_execute_api_endpoint"] = api.DisableExecuteApiEndpoint

				// Binary media types
				resource["binary_media_types"] = api.BinaryMediaTypes

				// Tags
				resource["tags"] = api.Tags

				// Get stages
				stagesResp, err := client.GetStages(ctx, &apigateway.GetStagesInput{
					RestApiId: api.Id,
				})
				if err == nil {
					stageNames := make([]string, 0, len(stagesResp.Item))
					hasLogging := false
					hasTracing := false
					hasClientCert := false
					hasWAF := false

					for _, stage := range stagesResp.Item {
						stageNames = append(stageNames, aws.ToString(stage.StageName))

						// Check stage settings
						if stage.AccessLogSettings != nil && aws.ToString(stage.AccessLogSettings.DestinationArn) != "" {
							hasLogging = true
						}
						if stage.TracingEnabled {
							hasTracing = true
						}
						if stage.ClientCertificateId != nil && aws.ToString(stage.ClientCertificateId) != "" {
							hasClientCert = true
						}
						if stage.WebAclArn != nil && aws.ToString(stage.WebAclArn) != "" {
							hasWAF = true
						}
					}

					resource["stages"] = stageNames
					resource["stage_count"] = len(stagesResp.Item)
					resource["has_logging"] = hasLogging
					resource["has_tracing"] = hasTracing
					resource["has_client_certificate"] = hasClientCert
					resource["has_waf"] = hasWAF
				}

				// Get resources count
				resourcesResp, err := client.GetResources(ctx, &apigateway.GetResourcesInput{
					RestApiId: api.Id,
				})
				if err == nil {
					resource["resource_count"] = len(resourcesResp.Items)
				}

				// Computed
				hasPolicy := aws.ToString(api.Policy) != ""
				resource["has_policy"] = hasPolicy

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

func containsEndpointType(types []string, t string) bool {
	for _, et := range types {
		if et == t {
			return true
		}
	}
	return false
}

// APIGatewayStage fetches API Gateway stages.
type APIGatewayStage struct {
	clientFactory APIGatewayClientFactory
	accountID     string
	regions       []string
}

// NewAPIGatewayStage creates a new APIGatewayStage resource.
func NewAPIGatewayStage(clientFactory APIGatewayClientFactory, accountID string, regions []string) *APIGatewayStage {
	return &APIGatewayStage{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *APIGatewayStage) Name() string {
	return "aws_apigateway_stage"
}

// Fetch retrieves all API Gateway stages across configured regions.
func (r *APIGatewayStage) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		// First get all REST APIs
		apiPaginator := apigateway.NewGetRestApisPaginator(client, &apigateway.GetRestApisInput{})

		for apiPaginator.HasMorePages() {
			apiPage, err := apiPaginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, api := range apiPage.Items {
				// Get stages for this API
				stagesResp, err := client.GetStages(ctx, &apigateway.GetStagesInput{
					RestApiId: api.Id,
				})
				if err != nil {
					continue
				}

				for _, stage := range stagesResp.Item {
					resource := make(core.Resource)
					resource["id"] = aws.ToString(api.Id) + "/" + aws.ToString(stage.StageName)
					resource["rest_api_id"] = aws.ToString(api.Id)
					resource["rest_api_name"] = aws.ToString(api.Name)
					resource["stage_name"] = aws.ToString(stage.StageName)
					resource["region"] = region

					resource["deployment_id"] = aws.ToString(stage.DeploymentId)
					resource["description"] = aws.ToString(stage.Description)

					// Caching
					resource["cache_cluster_enabled"] = stage.CacheClusterEnabled
					resource["cache_cluster_size"] = string(stage.CacheClusterSize)
					resource["cache_cluster_status"] = string(stage.CacheClusterStatus)

					// Logging
					if stage.AccessLogSettings != nil {
						resource["access_log_destination"] = aws.ToString(stage.AccessLogSettings.DestinationArn)
						resource["access_log_format"] = aws.ToString(stage.AccessLogSettings.Format)
						resource["has_access_logging"] = aws.ToString(stage.AccessLogSettings.DestinationArn) != ""
					} else {
						resource["has_access_logging"] = false
					}

					// Tracing
					resource["tracing_enabled"] = stage.TracingEnabled
					resource["has_xray_tracing"] = stage.TracingEnabled

					// Client certificate
					resource["client_certificate_id"] = aws.ToString(stage.ClientCertificateId)
					resource["has_client_certificate"] = stage.ClientCertificateId != nil && aws.ToString(stage.ClientCertificateId) != ""

					// Documentation version
					resource["documentation_version"] = aws.ToString(stage.DocumentationVersion)

					// Variables
					resource["stage_variables"] = stage.Variables
					resource["variable_count"] = len(stage.Variables)

					// Canary settings
					if stage.CanarySettings != nil {
						resource["canary_percent_traffic"] = stage.CanarySettings.PercentTraffic
						resource["canary_deployment_id"] = aws.ToString(stage.CanarySettings.DeploymentId)
						resource["has_canary"] = true
					} else {
						resource["has_canary"] = false
					}

					// Web ACL
					resource["web_acl_arn"] = aws.ToString(stage.WebAclArn)
					resource["has_waf"] = stage.WebAclArn != nil && aws.ToString(stage.WebAclArn) != ""

					// Timestamps
					if stage.CreatedDate != nil {
						resource["created_date"] = stage.CreatedDate.String()
					}
					if stage.LastUpdatedDate != nil {
						resource["last_updated_date"] = stage.LastUpdatedDate.String()
					}

					// Tags
					resource["tags"] = stage.Tags

					// Computed security baseline
					hasLogging, _ := resource["has_access_logging"].(bool) //nolint:errcheck // zero value ok
					hasWAF, _ := resource["has_waf"].(bool)                //nolint:errcheck // zero value ok
					resource["meets_security_baseline"] = hasLogging && hasWAF

					resources = append(resources, resource)
				}
			}
		}
	}

	return resources, nil
}

// APIGatewayV2API fetches API Gateway HTTP/WebSocket APIs.
type APIGatewayV2API struct {
	clientFactory APIGatewayV2ClientFactory
	accountID     string
	regions       []string
}

// NewAPIGatewayV2API creates a new APIGatewayV2API resource.
func NewAPIGatewayV2API(clientFactory APIGatewayV2ClientFactory, accountID string, regions []string) *APIGatewayV2API {
	return &APIGatewayV2API{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *APIGatewayV2API) Name() string {
	return "aws_apigatewayv2_api"
}

// Fetch retrieves all API Gateway V2 APIs across configured regions.
func (r *APIGatewayV2API) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		resp, err := client.GetApis(ctx, &apigatewayv2.GetApisInput{})
		if err != nil {
			continue
		}

		for _, api := range resp.Items {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(api.ApiId)
			resource["name"] = aws.ToString(api.Name)
			resource["region"] = region

			resource["api_endpoint"] = aws.ToString(api.ApiEndpoint)
			resource["description"] = aws.ToString(api.Description)
			resource["protocol_type"] = string(api.ProtocolType)
			resource["route_selection_expression"] = aws.ToString(api.RouteSelectionExpression)
			resource["api_key_selection_expression"] = aws.ToString(api.ApiKeySelectionExpression)
			resource["version"] = aws.ToString(api.Version)

			// Disable execute API endpoint
			resource["disable_execute_api_endpoint"] = api.DisableExecuteApiEndpoint

			// API type
			resource["is_http_api"] = api.ProtocolType == "HTTP"
			resource["is_websocket_api"] = api.ProtocolType == "WEBSOCKET"

			// CORS
			if api.CorsConfiguration != nil {
				resource["has_cors"] = true
				resource["cors_allow_origins"] = api.CorsConfiguration.AllowOrigins
				resource["cors_allow_methods"] = api.CorsConfiguration.AllowMethods
				resource["cors_allow_headers"] = api.CorsConfiguration.AllowHeaders
				resource["cors_expose_headers"] = api.CorsConfiguration.ExposeHeaders
				resource["cors_max_age"] = api.CorsConfiguration.MaxAge
				resource["cors_allow_credentials"] = api.CorsConfiguration.AllowCredentials

				// Check for wildcard origin
				hasWildcardOrigin := false
				for _, origin := range api.CorsConfiguration.AllowOrigins {
					if origin == "*" {
						hasWildcardOrigin = true
						break
					}
				}
				resource["cors_has_wildcard_origin"] = hasWildcardOrigin
			} else {
				resource["has_cors"] = false
			}

			// Created date
			if api.CreatedDate != nil {
				resource["created_date"] = api.CreatedDate.String()
			}

			// Tags
			resource["tags"] = api.Tags

			// Get stages
			stagesResp, err := client.GetStages(ctx, &apigatewayv2.GetStagesInput{
				ApiId: api.ApiId,
			})
			if err == nil {
				stageNames := make([]string, 0, len(stagesResp.Items))
				hasLogging := false
				hasThrottling := false

				for _, stage := range stagesResp.Items {
					stageNames = append(stageNames, aws.ToString(stage.StageName))

					if stage.AccessLogSettings != nil && aws.ToString(stage.AccessLogSettings.DestinationArn) != "" {
						hasLogging = true
					}
					if stage.DefaultRouteSettings != nil && stage.DefaultRouteSettings.ThrottlingBurstLimit != nil && *stage.DefaultRouteSettings.ThrottlingBurstLimit > 0 {
						hasThrottling = true
					}
				}

				resource["stages"] = stageNames
				resource["stage_count"] = len(stagesResp.Items)
				resource["has_logging"] = hasLogging
				resource["has_throttling"] = hasThrottling
			}

			// Get routes count
			routesResp, err := client.GetRoutes(ctx, &apigatewayv2.GetRoutesInput{
				ApiId: api.ApiId,
			})
			if err == nil {
				resource["route_count"] = len(routesResp.Items)

				// Check for authorization
				hasAuthorization := false
				for _, route := range routesResp.Items {
					if route.AuthorizationType != "NONE" {
						hasAuthorization = true
						break
					}
				}
				resource["has_authorization"] = hasAuthorization
			}

			// Get integrations count
			integrationsResp, err := client.GetIntegrations(ctx, &apigatewayv2.GetIntegrationsInput{
				ApiId: api.ApiId,
			})
			if err == nil {
				resource["integration_count"] = len(integrationsResp.Items)
			}

			// Get authorizers
			authorizersResp, err := client.GetAuthorizers(ctx, &apigatewayv2.GetAuthorizersInput{
				ApiId: api.ApiId,
			})
			if err == nil {
				authorizerNames := make([]string, 0, len(authorizersResp.Items))
				for _, auth := range authorizersResp.Items {
					authorizerNames = append(authorizerNames, aws.ToString(auth.Name))
				}
				resource["authorizers"] = authorizerNames
				resource["authorizer_count"] = len(authorizersResp.Items)
				resource["has_authorizers"] = len(authorizersResp.Items) > 0
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
