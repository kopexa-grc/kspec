// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"go.uber.org/mock/gomock"

	"github.com/kopexa-grc/kspec/core"
)

func TestELBLoadBalancerResource_Name(t *testing.T) {
	r := &ELBLoadBalancerResource{}
	if got := r.Name(); got != "aws_elb_loadbalancer" {
		t.Errorf("Name() = %v, want aws_elb_loadbalancer", got)
	}
}

func TestELBLoadBalancerResource_Fetch_EmptyLoadBalancers(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockELBv2API(ctrl)

	mock.EXPECT().
		DescribeLoadBalancers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{LoadBalancers: []types.LoadBalancer{}}, nil)

	r := &ELBLoadBalancerResource{
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

func TestELBLoadBalancerResource_Fetch_ALB(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockELBv2API(ctrl)

	lbName := "my-alb"
	lbArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/12345"

	mock.EXPECT().
		DescribeLoadBalancers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{
				{
					LoadBalancerName:      aws.String(lbName),
					LoadBalancerArn:       aws.String(lbArn),
					DNSName:               aws.String("my-alb-123456.us-east-1.elb.amazonaws.com"),
					Type:                  types.LoadBalancerTypeEnumApplication,
					Scheme:                types.LoadBalancerSchemeEnumInternetFacing,
					State:                 &types.LoadBalancerState{Code: types.LoadBalancerStateEnumActive},
					VpcId:                 aws.String("vpc-12345"),
					CanonicalHostedZoneId: aws.String("Z35SXDOTRQ7X7K"),
					IpAddressType:         types.IpAddressTypeIpv4,
					AvailabilityZones: []types.AvailabilityZone{
						{ZoneName: aws.String("us-east-1a"), SubnetId: aws.String("subnet-1")},
						{ZoneName: aws.String("us-east-1b"), SubnetId: aws.String("subnet-2")},
					},
					SecurityGroups: []string{"sg-alb"},
				},
			},
		}, nil)

	// Attributes
	mock.EXPECT().
		DescribeLoadBalancerAttributes(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{
			Attributes: []types.LoadBalancerAttribute{
				{Key: aws.String("access_logs.s3.enabled"), Value: aws.String("true")},
				{Key: aws.String("access_logs.s3.bucket"), Value: aws.String("my-logs-bucket")},
				{Key: aws.String("deletion_protection.enabled"), Value: aws.String("true")},
				{Key: aws.String("routing.http.drop_invalid_header_fields.enabled"), Value: aws.String("true")},
				{Key: aws.String("routing.http2.enabled"), Value: aws.String("true")},
			},
		}, nil)

	// Listeners
	mock.EXPECT().
		DescribeListeners(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeListenersOutput{
			Listeners: []types.Listener{
				{
					Port:        aws.Int32(443),
					Protocol:    types.ProtocolEnumHttps,
					ListenerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/my-alb/12345/67890"),
				},
			},
		}, nil)

	// Tags
	mock.EXPECT().
		DescribeTags(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeTagsOutput{
			TagDescriptions: []types.TagDescription{
				{
					ResourceArn: aws.String(lbArn),
					Tags: []types.Tag{
						{Key: aws.String("Environment"), Value: aws.String("production")},
					},
				},
			},
		}, nil)

	r := &ELBLoadBalancerResource{
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
	if res["name"] != lbName {
		t.Errorf("name = %v, want %v", res["name"], lbName)
	}
	if res["is_application"] != true {
		t.Errorf("is_application = %v, want true", res["is_application"])
	}
	if res["is_internet_facing"] != true {
		t.Errorf("is_internet_facing = %v, want true", res["is_internet_facing"])
	}
	if res["is_active"] != true {
		t.Errorf("is_active = %v, want true", res["is_active"])
	}
	if res["has_access_logs"] != true {
		t.Errorf("has_access_logs = %v, want true", res["has_access_logs"])
	}
	if res["has_deletion_protection"] != true {
		t.Errorf("has_deletion_protection = %v, want true", res["has_deletion_protection"])
	}
	if res["has_https_listener"] != true {
		t.Errorf("has_https_listener = %v, want true", res["has_https_listener"])
	}
	if res["drop_invalid_headers"] != true {
		t.Errorf("drop_invalid_headers = %v, want true", res["drop_invalid_headers"])
	}
}

func TestELBLoadBalancerResource_Fetch_NLB(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockELBv2API(ctrl)

	lbName := "my-nlb"
	lbArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/net/my-nlb/12345"

	mock.EXPECT().
		DescribeLoadBalancers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{
				{
					LoadBalancerName: aws.String(lbName),
					LoadBalancerArn:  aws.String(lbArn),
					Type:             types.LoadBalancerTypeEnumNetwork,
					Scheme:           types.LoadBalancerSchemeEnumInternal,
					State:            &types.LoadBalancerState{Code: types.LoadBalancerStateEnumActive},
					AvailabilityZones: []types.AvailabilityZone{
						{ZoneName: aws.String("us-east-1a"), SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}, nil)

	mock.EXPECT().DescribeLoadBalancerAttributes(gomock.Any(), gomock.Any()).Return(&elasticloadbalancingv2.DescribeLoadBalancerAttributesOutput{}, nil)
	mock.EXPECT().DescribeListeners(gomock.Any(), gomock.Any()).Return(&elasticloadbalancingv2.DescribeListenersOutput{}, nil)
	mock.EXPECT().DescribeTags(gomock.Any(), gomock.Any()).Return(&elasticloadbalancingv2.DescribeTagsOutput{}, nil)

	r := &ELBLoadBalancerResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["is_network"] != true {
		t.Errorf("is_network = %v, want true", res["is_network"])
	}
	if res["is_internal"] != true {
		t.Errorf("is_internal = %v, want true", res["is_internal"])
	}
}

func TestELBTargetGroupResource_Name(t *testing.T) {
	r := &ELBTargetGroupResource{}
	if got := r.Name(); got != "aws_elb_targetgroup" {
		t.Errorf("Name() = %v, want aws_elb_targetgroup", got)
	}
}

func TestELBTargetGroupResource_Fetch_TargetGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockELBv2API(ctrl)

	tgName := "my-target-group"
	tgArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:targetgroup/my-target-group/12345"

	mock.EXPECT().
		DescribeTargetGroups(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeTargetGroupsOutput{
			TargetGroups: []types.TargetGroup{
				{
					TargetGroupName:            aws.String(tgName),
					TargetGroupArn:             aws.String(tgArn),
					Protocol:                   types.ProtocolEnumHttps,
					Port:                       aws.Int32(443),
					TargetType:                 types.TargetTypeEnumInstance,
					VpcId:                      aws.String("vpc-12345"),
					HealthCheckEnabled:         aws.Bool(true),
					HealthCheckProtocol:        types.ProtocolEnumHttps,
					HealthCheckPort:            aws.String("443"),
					HealthCheckPath:            aws.String("/health"),
					HealthCheckIntervalSeconds: aws.Int32(30),
					HealthCheckTimeoutSeconds:  aws.Int32(5),
					HealthyThresholdCount:      aws.Int32(2),
					UnhealthyThresholdCount:    aws.Int32(2),
					Matcher:                    &types.Matcher{HttpCode: aws.String("200")},
					LoadBalancerArns:           []string{"arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/12345"},
				},
			},
		}, nil)

	// Target health
	mock.EXPECT().
		DescribeTargetHealth(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeTargetHealthOutput{
			TargetHealthDescriptions: []types.TargetHealthDescription{
				{TargetHealth: &types.TargetHealth{State: types.TargetHealthStateEnumHealthy}},
				{TargetHealth: &types.TargetHealth{State: types.TargetHealthStateEnumHealthy}},
			},
		}, nil)

	// Attributes
	mock.EXPECT().
		DescribeTargetGroupAttributes(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeTargetGroupAttributesOutput{
			Attributes: []types.TargetGroupAttribute{
				{Key: aws.String("stickiness.enabled"), Value: aws.String("true")},
				{Key: aws.String("stickiness.type"), Value: aws.String("lb_cookie")},
			},
		}, nil)

	// Tags
	mock.EXPECT().
		DescribeTags(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeTagsOutput{}, nil)

	r := &ELBTargetGroupResource{
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
	if res["name"] != tgName {
		t.Errorf("name = %v, want %v", res["name"], tgName)
	}
	if res["is_attached"] != true {
		t.Errorf("is_attached = %v, want true", res["is_attached"])
	}
	if res["is_instance_target"] != true {
		t.Errorf("is_instance_target = %v, want true", res["is_instance_target"])
	}
	if res["healthy_target_count"] != 2 {
		t.Errorf("healthy_target_count = %v, want 2", res["healthy_target_count"])
	}
	if res["all_targets_healthy"] != true {
		t.Errorf("all_targets_healthy = %v, want true", res["all_targets_healthy"])
	}
	if res["stickiness_enabled"] != true {
		t.Errorf("stickiness_enabled = %v, want true", res["stickiness_enabled"])
	}
}

func TestELBListenerResource_Name(t *testing.T) {
	r := &ELBListenerResource{}
	if got := r.Name(); got != "aws_elb_listener" {
		t.Errorf("Name() = %v, want aws_elb_listener", got)
	}
}

func TestELBListenerResource_Fetch_HTTPSListener(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockELBv2API(ctrl)

	lbArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/my-alb/12345"
	listenerArn := "arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/my-alb/12345/67890"

	mock.EXPECT().
		DescribeLoadBalancers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{
				{
					LoadBalancerName: aws.String("my-alb"),
					LoadBalancerArn:  aws.String(lbArn),
					Type:             types.LoadBalancerTypeEnumApplication,
				},
			},
		}, nil)

	mock.EXPECT().
		DescribeListeners(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeListenersOutput{
			Listeners: []types.Listener{
				{
					ListenerArn: aws.String(listenerArn),
					Protocol:    types.ProtocolEnumHttps,
					Port:        aws.Int32(443),
					Certificates: []types.Certificate{
						{CertificateArn: aws.String("arn:aws:acm:us-east-1:123456789012:certificate/12345")},
					},
					SslPolicy: aws.String("ELBSecurityPolicy-TLS13-1-2-2021-06"),
					DefaultActions: []types.Action{
						{Type: types.ActionTypeEnumForward},
					},
				},
			},
		}, nil)

	r := &ELBListenerResource{
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
	if res["is_https"] != true {
		t.Errorf("is_https = %v, want true", res["is_https"])
	}
	if res["is_secure"] != true {
		t.Errorf("is_secure = %v, want true", res["is_secure"])
	}
	if res["has_certificate"] != true {
		t.Errorf("has_certificate = %v, want true", res["has_certificate"])
	}
	if res["has_modern_ssl_policy"] != true {
		t.Errorf("has_modern_ssl_policy = %v, want true for TLS13", res["has_modern_ssl_policy"])
	}
}

func TestELBListenerResource_Fetch_HTTPListener(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockELBv2API(ctrl)

	mock.EXPECT().
		DescribeLoadBalancers(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeLoadBalancersOutput{
			LoadBalancers: []types.LoadBalancer{
				{
					LoadBalancerName: aws.String("http-alb"),
					LoadBalancerArn:  aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:loadbalancer/app/http-alb/12345"),
					Type:             types.LoadBalancerTypeEnumApplication,
				},
			},
		}, nil)

	mock.EXPECT().
		DescribeListeners(gomock.Any(), gomock.Any()).
		Return(&elasticloadbalancingv2.DescribeListenersOutput{
			Listeners: []types.Listener{
				{
					ListenerArn: aws.String("arn:aws:elasticloadbalancing:us-east-1:123456789012:listener/app/http-alb/12345/http"),
					Protocol:    types.ProtocolEnumHttp,
					Port:        aws.Int32(80),
				},
			},
		}, nil)

	r := &ELBListenerResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["is_http"] != true {
		t.Errorf("is_http = %v, want true", res["is_http"])
	}
	if res["is_secure"] != false {
		t.Errorf("is_secure = %v, want false for HTTP", res["is_secure"])
	}
	if res["has_certificate"] != false {
		t.Errorf("has_certificate = %v, want false", res["has_certificate"])
	}
}
