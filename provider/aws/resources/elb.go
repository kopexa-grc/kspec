// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ptr"
)

// ELBClient is the interface for ELB operations.
//
//nolint:dupl // AWS SDK client interfaces have similar structures by design
type ELBClient interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
	DescribeLoadBalancerAttributes(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancerAttributesInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput, error)
	DescribeListeners(ctx context.Context, params *elasticloadbalancingv2.DescribeListenersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeListenersOutput, error)
	DescribeTargetGroups(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupsOutput, error)
	DescribeTargetGroupAttributes(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetGroupAttributesInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetGroupAttributesOutput, error)
	DescribeTargetHealth(ctx context.Context, params *elasticloadbalancingv2.DescribeTargetHealthInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTargetHealthOutput, error)
	DescribeTags(ctx context.Context, params *elasticloadbalancingv2.DescribeTagsInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeTagsOutput, error)
}

// ELBClientFactory creates ELB clients for specific regions.
type ELBClientFactory func(region string) ELBClient

// ELBLoadBalancer fetches Application/Network Load Balancers.
type ELBLoadBalancer struct {
	clientFactory ELBClientFactory
	accountID     string
	regions       []string
}

// NewELBLoadBalancer creates a new ELBLoadBalancer resource.
func NewELBLoadBalancer(clientFactory ELBClientFactory, accountID string, regions []string) *ELBLoadBalancer {
	return &ELBLoadBalancer{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ELBLoadBalancer) Name() string {
	return "aws_elb_loadbalancer"
}

// Fetch retrieves all ALBs/NLBs across configured regions.
func (r *ELBLoadBalancer) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(client, &elasticloadbalancingv2.DescribeLoadBalancersInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, lb := range page.LoadBalancers {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(lb.LoadBalancerName)
				resource["arn"] = aws.ToString(lb.LoadBalancerArn)
				resource["name"] = aws.ToString(lb.LoadBalancerName)
				resource["region"] = region

				resource["dns_name"] = aws.ToString(lb.DNSName)
				resource["type"] = string(lb.Type)
				resource["scheme"] = string(lb.Scheme)
				resource["state"] = string(lb.State.Code)
				resource["vpc_id"] = aws.ToString(lb.VpcId)
				resource["canonical_hosted_zone_id"] = aws.ToString(lb.CanonicalHostedZoneId)

				// IP address type
				resource["ip_address_type"] = string(lb.IpAddressType)

				// Availability zones
				azs := make([]string, 0, len(lb.AvailabilityZones))
				subnets := make([]string, 0, len(lb.AvailabilityZones))
				for _, az := range lb.AvailabilityZones {
					azs = append(azs, aws.ToString(az.ZoneName))
					subnets = append(subnets, aws.ToString(az.SubnetId))
				}
				resource["availability_zones"] = azs
				resource["subnets"] = subnets

				// Security groups (ALB only)
				resource["security_groups"] = lb.SecurityGroups

				// Computed
				resource["is_internal"] = lb.Scheme == types.LoadBalancerSchemeEnumInternal
				resource["is_internet_facing"] = lb.Scheme == types.LoadBalancerSchemeEnumInternetFacing
				resource["is_application"] = lb.Type == types.LoadBalancerTypeEnumApplication
				resource["is_network"] = lb.Type == types.LoadBalancerTypeEnumNetwork
				resource["is_gateway"] = lb.Type == types.LoadBalancerTypeEnumGateway
				resource["is_active"] = lb.State.Code == types.LoadBalancerStateEnumActive

				// Get attributes
				attrsResp, err := client.DescribeLoadBalancerAttributes(ctx, &elasticloadbalancingv2.DescribeLoadBalancerAttributesInput{
					LoadBalancerArn: lb.LoadBalancerArn,
				})
				if err == nil {
					for _, attr := range attrsResp.Attributes {
						key := aws.ToString(attr.Key)
						value := aws.ToString(attr.Value)

						switch key {
						case "access_logs.s3.enabled":
							accessLogs := value == valueTrue
							resource["access_logs_enabled"] = accessLogs
							resource["has_access_logs"] = accessLogs
						case "access_logs.s3.bucket":
							resource["access_logs_bucket"] = value
						case "access_logs.s3.prefix":
							resource["access_logs_prefix"] = value
						case "deletion_protection.enabled":
							deletionProtection := value == valueTrue
							resource["deletion_protection"] = deletionProtection
							resource["has_deletion_protection"] = deletionProtection
						case "idle_timeout.timeout_seconds":
							resource["idle_timeout"] = value
						case "routing.http.drop_invalid_header_fields.enabled":
							resource["drop_invalid_headers"] = value == valueTrue
						case "routing.http2.enabled":
							resource["http2_enabled"] = value == valueTrue
						case "routing.http.desync_mitigation_mode":
							resource["desync_mitigation_mode"] = value
						case "load_balancing.cross_zone.enabled":
							resource["cross_zone_enabled"] = value == valueTrue
						case "waf.fail_open.enabled":
							resource["waf_fail_open"] = value == valueTrue
						}
					}
				}

				// Get listeners
				listenersResp, err := client.DescribeListeners(ctx, &elasticloadbalancingv2.DescribeListenersInput{
					LoadBalancerArn: lb.LoadBalancerArn,
				})
				if err == nil {
					listenerPorts := make([]int32, 0, len(listenersResp.Listeners))
					hasHTTPS := false
					hasHTTP := false
					for _, listener := range listenersResp.Listeners {
						listenerPorts = append(listenerPorts, ptr.Deref(listener.Port, 0))
						if listener.Protocol == types.ProtocolEnumHttps || listener.Protocol == types.ProtocolEnumTls {
							hasHTTPS = true
						}
						if listener.Protocol == types.ProtocolEnumHttp {
							hasHTTP = true
						}
					}
					resource["listener_ports"] = listenerPorts
					resource["listener_count"] = len(listenersResp.Listeners)
					resource["has_https_listener"] = hasHTTPS
					resource["has_http_listener"] = hasHTTP
					resource["has_secure_listener"] = hasHTTPS
				}

				// Get tags
				tagsResp, err := client.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
					ResourceArns: []string{aws.ToString(lb.LoadBalancerArn)},
				})
				if err == nil && len(tagsResp.TagDescriptions) > 0 {
					tags := make(map[string]string)
					for _, tag := range tagsResp.TagDescriptions[0].Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// ELBTargetGroup fetches ELB target groups.
type ELBTargetGroup struct {
	clientFactory ELBClientFactory
	accountID     string
	regions       []string
}

// NewELBTargetGroup creates a new ELBTargetGroup resource.
func NewELBTargetGroup(clientFactory ELBClientFactory, accountID string, regions []string) *ELBTargetGroup {
	return &ELBTargetGroup{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ELBTargetGroup) Name() string {
	return "aws_elb_targetgroup"
}

// Fetch retrieves all target groups across configured regions.
func (r *ELBTargetGroup) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(client, &elasticloadbalancingv2.DescribeTargetGroupsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, tg := range page.TargetGroups {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(tg.TargetGroupName)
				resource["arn"] = aws.ToString(tg.TargetGroupArn)
				resource["name"] = aws.ToString(tg.TargetGroupName)
				resource["region"] = region

				resource["protocol"] = string(tg.Protocol)
				resource["port"] = tg.Port
				resource["target_type"] = string(tg.TargetType)
				resource["vpc_id"] = aws.ToString(tg.VpcId)

				// Health check
				resource["health_check_enabled"] = ptr.Deref(tg.HealthCheckEnabled, false)
				resource["health_check_protocol"] = string(tg.HealthCheckProtocol)
				resource["health_check_port"] = aws.ToString(tg.HealthCheckPort)
				resource["health_check_path"] = aws.ToString(tg.HealthCheckPath)
				resource["health_check_interval"] = tg.HealthCheckIntervalSeconds
				resource["health_check_timeout"] = tg.HealthCheckTimeoutSeconds
				resource["healthy_threshold"] = tg.HealthyThresholdCount
				resource["unhealthy_threshold"] = tg.UnhealthyThresholdCount
				resource["health_check_matcher"] = aws.ToString(tg.Matcher.HttpCode)

				// Load balancer associations
				resource["load_balancer_arns"] = tg.LoadBalancerArns
				resource["is_attached"] = len(tg.LoadBalancerArns) > 0

				// IP address type
				resource["ip_address_type"] = string(tg.IpAddressType)

				// Protocol version (for HTTP/HTTPS)
				resource["protocol_version"] = aws.ToString(tg.ProtocolVersion)

				// Computed
				resource["is_instance_target"] = tg.TargetType == types.TargetTypeEnumInstance
				resource["is_ip_target"] = tg.TargetType == types.TargetTypeEnumIp
				resource["is_lambda_target"] = tg.TargetType == types.TargetTypeEnumLambda
				resource["is_alb_target"] = tg.TargetType == types.TargetTypeEnumAlb

				// Get target health
				healthResp, err := client.DescribeTargetHealth(ctx, &elasticloadbalancingv2.DescribeTargetHealthInput{
					TargetGroupArn: tg.TargetGroupArn,
				})
				if err == nil {
					healthyCount := 0
					unhealthyCount := 0
					for _, th := range healthResp.TargetHealthDescriptions {
						switch th.TargetHealth.State { //nolint:exhaustive // only care about healthy/unhealthy
						case types.TargetHealthStateEnumHealthy:
							healthyCount++
						case types.TargetHealthStateEnumUnhealthy:
							unhealthyCount++
						default:
							// Other states (initial, draining, etc.)
						}
					}
					resource["target_count"] = len(healthResp.TargetHealthDescriptions)
					resource["healthy_target_count"] = healthyCount
					resource["unhealthy_target_count"] = unhealthyCount
					resource["has_healthy_targets"] = healthyCount > 0
					resource["all_targets_healthy"] = healthyCount == len(healthResp.TargetHealthDescriptions) && len(healthResp.TargetHealthDescriptions) > 0
				}

				// Get attributes
				attrsResp, err := client.DescribeTargetGroupAttributes(ctx, &elasticloadbalancingv2.DescribeTargetGroupAttributesInput{
					TargetGroupArn: tg.TargetGroupArn,
				})
				if err == nil {
					for _, attr := range attrsResp.Attributes {
						key := aws.ToString(attr.Key)
						value := aws.ToString(attr.Value)

						switch key {
						case "stickiness.enabled":
							resource["stickiness_enabled"] = value == valueTrue
						case "stickiness.type":
							resource["stickiness_type"] = value
						case "deregistration_delay.timeout_seconds":
							resource["deregistration_delay"] = value
						case "slow_start.duration_seconds":
							resource["slow_start_duration"] = value
						case "load_balancing.algorithm.type":
							resource["load_balancing_algorithm"] = value
						}
					}
				}

				// Get tags
				tagsResp, err := client.DescribeTags(ctx, &elasticloadbalancingv2.DescribeTagsInput{
					ResourceArns: []string{aws.ToString(tg.TargetGroupArn)},
				})
				if err == nil && len(tagsResp.TagDescriptions) > 0 {
					tags := make(map[string]string)
					for _, tag := range tagsResp.TagDescriptions[0].Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// ELBListener fetches ELB listeners.
type ELBListener struct {
	clientFactory ELBClientFactory
	accountID     string
	regions       []string
}

// NewELBListener creates a new ELBListener resource.
func NewELBListener(clientFactory ELBClientFactory, accountID string, regions []string) *ELBListener {
	return &ELBListener{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ELBListener) Name() string {
	return "aws_elb_listener"
}

// Fetch retrieves all listeners across configured regions.
func (r *ELBListener) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		// First get all load balancers
		lbPaginator := elasticloadbalancingv2.NewDescribeLoadBalancersPaginator(client, &elasticloadbalancingv2.DescribeLoadBalancersInput{})

		for lbPaginator.HasMorePages() {
			lbPage, err := lbPaginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, lb := range lbPage.LoadBalancers {
				// Get listeners for this load balancer
				listenersResp, err := client.DescribeListeners(ctx, &elasticloadbalancingv2.DescribeListenersInput{
					LoadBalancerArn: lb.LoadBalancerArn,
				})
				if err != nil {
					continue
				}

				for _, listener := range listenersResp.Listeners {
					resource := make(core.Resource)
					resource["id"] = aws.ToString(listener.ListenerArn)
					resource["arn"] = aws.ToString(listener.ListenerArn)
					resource["load_balancer_arn"] = aws.ToString(lb.LoadBalancerArn)
					resource["load_balancer_name"] = aws.ToString(lb.LoadBalancerName)
					resource["region"] = region

					resource["protocol"] = string(listener.Protocol)
					resource["port"] = listener.Port

					// SSL/TLS
					if len(listener.Certificates) > 0 {
						resource["certificate_arn"] = aws.ToString(listener.Certificates[0].CertificateArn)
						resource["has_certificate"] = true
					} else {
						resource["has_certificate"] = false
					}
					resource["ssl_policy"] = aws.ToString(listener.SslPolicy)

					// Check for secure SSL policy
					sslPolicy := aws.ToString(listener.SslPolicy)
					resource["has_secure_ssl_policy"] = strings.Contains(sslPolicy, "TLS-1-2") || strings.Contains(sslPolicy, "TLS13")
					resource["has_modern_ssl_policy"] = strings.Contains(sslPolicy, "TLS13")

					// Default actions
					defaultActionTypes := make([]string, 0, len(listener.DefaultActions))
					for _, action := range listener.DefaultActions {
						defaultActionTypes = append(defaultActionTypes, string(action.Type))
					}
					resource["default_action_types"] = defaultActionTypes
					resource["default_action_count"] = len(listener.DefaultActions)

					// Computed
					resource["is_https"] = listener.Protocol == types.ProtocolEnumHttps
					resource["is_http"] = listener.Protocol == types.ProtocolEnumHttp
					resource["is_tls"] = listener.Protocol == types.ProtocolEnumTls
					resource["is_tcp"] = listener.Protocol == types.ProtocolEnumTcp
					resource["is_secure"] = listener.Protocol == types.ProtocolEnumHttps || listener.Protocol == types.ProtocolEnumTls

					// Mutual authentication
					if listener.MutualAuthentication != nil {
						resource["mutual_auth_mode"] = aws.ToString(listener.MutualAuthentication.Mode)
						resource["has_mutual_auth"] = listener.MutualAuthentication.Mode != nil && aws.ToString(listener.MutualAuthentication.Mode) != ""
					} else {
						resource["has_mutual_auth"] = false
					}

					resources = append(resources, resource)
				}
			}
		}
	}

	return resources, nil
}
