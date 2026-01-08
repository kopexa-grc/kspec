package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/kopexa-grc/kspec/core"
)

// VPCResource fetches VPCs.
type VPCResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *VPCResource) Name() string {
	return "aws_vpc"
}

// Fetch retrieves all VPCs across configured regions.
func (r *VPCResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := ec2.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := ec2.NewDescribeVpcsPaginator(client, &ec2.DescribeVpcsInput{})

		// Get flow logs for this region (ignore error - empty map on failure is acceptable)
		flowLogsResp, _ := client.DescribeFlowLogs(ctx, &ec2.DescribeFlowLogsInput{}) //nolint:errcheck // intentionally ignoring error
		vpcFlowLogs := make(map[string]bool)
		if flowLogsResp != nil {
			for _, fl := range flowLogsResp.FlowLogs {
				if fl.ResourceId != nil {
					vpcFlowLogs[aws.ToString(fl.ResourceId)] = true
				}
			}
		}

		// Get internet gateways (ignore error - empty map on failure is acceptable)
		igwResp, _ := client.DescribeInternetGateways(ctx, &ec2.DescribeInternetGatewaysInput{}) //nolint:errcheck // intentionally ignoring error
		vpcIGWs := make(map[string]bool)
		if igwResp != nil {
			for _, igw := range igwResp.InternetGateways {
				for _, attachment := range igw.Attachments {
					if attachment.VpcId != nil {
						vpcIGWs[aws.ToString(attachment.VpcId)] = true
					}
				}
			}
		}

		// Get NAT gateways (ignore error - empty map on failure is acceptable)
		natResp, _ := client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{}) //nolint:errcheck // intentionally ignoring error
		vpcNATs := make(map[string]bool)
		if natResp != nil {
			for _, nat := range natResp.NatGateways {
				if nat.VpcId != nil {
					vpcNATs[aws.ToString(nat.VpcId)] = true
				}
			}
		}

		// Get subnets for counting (ignore error - empty map on failure is acceptable)
		subnetsResp, _ := client.DescribeSubnets(ctx, &ec2.DescribeSubnetsInput{}) //nolint:errcheck // intentionally ignoring error
		vpcSubnetCounts := make(map[string]int)
		if subnetsResp != nil {
			for _, subnet := range subnetsResp.Subnets {
				if subnet.VpcId != nil {
					vpcSubnetCounts[aws.ToString(subnet.VpcId)]++
				}
			}
		}

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, vpc := range page.Vpcs {
				vpcID := aws.ToString(vpc.VpcId)
				resource := make(core.Resource)
				resource["id"] = vpcID
				resource["region"] = region
				resource["arn"] = "arn:aws:ec2:" + region + ":" + r.conn.accountID + ":vpc/" + vpcID

				// Get name from tags
				for _, tag := range vpc.Tags {
					if aws.ToString(tag.Key) == "Name" {
						resource["name"] = aws.ToString(tag.Value)
						break
					}
				}

				resource["cidr_block"] = aws.ToString(vpc.CidrBlock)
				resource["state"] = string(vpc.State)
				resource["owner_id"] = aws.ToString(vpc.OwnerId)
				resource["instance_tenancy"] = string(vpc.InstanceTenancy)

				// Default VPC
				resource["is_default"] = vpc.IsDefault != nil && *vpc.IsDefault

				// DNS settings
				resource["enable_dns_support"] = true // Default, need to check attributes
				resource["enable_dns_hostnames"] = false

				// Get VPC attributes
				dnsSupport, err := client.DescribeVpcAttribute(ctx, &ec2.DescribeVpcAttributeInput{
					VpcId:     vpc.VpcId,
					Attribute: types.VpcAttributeNameEnableDnsSupport,
				})
				if err == nil && dnsSupport.EnableDnsSupport != nil {
					resource["enable_dns_support"] = dnsSupport.EnableDnsSupport.Value != nil && *dnsSupport.EnableDnsSupport.Value
				}

				dnsHostnames, err := client.DescribeVpcAttribute(ctx, &ec2.DescribeVpcAttributeInput{
					VpcId:     vpc.VpcId,
					Attribute: types.VpcAttributeNameEnableDnsHostnames,
				})
				if err == nil && dnsHostnames.EnableDnsHostnames != nil {
					resource["enable_dns_hostnames"] = dnsHostnames.EnableDnsHostnames.Value != nil && *dnsHostnames.EnableDnsHostnames.Value
				}

				// CIDR blocks
				var cidrBlocks []string
				for _, assoc := range vpc.CidrBlockAssociationSet {
					if assoc.CidrBlock != nil {
						cidrBlocks = append(cidrBlocks, aws.ToString(assoc.CidrBlock))
					}
				}
				resource["cidr_blocks"] = cidrBlocks

				var ipv6Blocks []string
				for _, assoc := range vpc.Ipv6CidrBlockAssociationSet {
					if assoc.Ipv6CidrBlock != nil {
						ipv6Blocks = append(ipv6Blocks, aws.ToString(assoc.Ipv6CidrBlock))
					}
				}
				resource["ipv6_cidr_blocks"] = ipv6Blocks
				resource["has_ipv6"] = len(ipv6Blocks) > 0

				// Computed fields using pre-fetched data
				resource["has_flow_logs"] = vpcFlowLogs[vpcID]
				resource["has_internet_gateway"] = vpcIGWs[vpcID]
				resource["has_nat_gateway"] = vpcNATs[vpcID]
				resource["subnet_count"] = vpcSubnetCounts[vpcID]

				// Tags
				tags := make(map[string]string)
				for _, tag := range vpc.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// VPCSubnetResource fetches VPC subnets.
type VPCSubnetResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *VPCSubnetResource) Name() string {
	return "aws_vpc_subnet"
}

// Fetch retrieves all VPC subnets across configured regions.
func (r *VPCSubnetResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := ec2.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := ec2.NewDescribeSubnetsPaginator(client, &ec2.DescribeSubnetsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, subnet := range page.Subnets {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(subnet.SubnetId)
				resource["region"] = region
				resource["arn"] = aws.ToString(subnet.SubnetArn)

				// Get name from tags
				for _, tag := range subnet.Tags {
					if aws.ToString(tag.Key) == "Name" {
						resource["name"] = aws.ToString(tag.Value)
						break
					}
				}

				resource["vpc_id"] = aws.ToString(subnet.VpcId)
				resource["cidr_block"] = aws.ToString(subnet.CidrBlock)
				resource["availability_zone"] = aws.ToString(subnet.AvailabilityZone)
				resource["availability_zone_id"] = aws.ToString(subnet.AvailabilityZoneId)
				resource["state"] = string(subnet.State)
				resource["owner_id"] = aws.ToString(subnet.OwnerId)

				// IP counts
				resource["available_ip_count"] = subnet.AvailableIpAddressCount

				// Default subnet
				resource["is_default"] = subnet.DefaultForAz != nil && *subnet.DefaultForAz

				// Auto-assign public IP
				resource["map_public_ip_on_launch"] = subnet.MapPublicIpOnLaunch != nil && *subnet.MapPublicIpOnLaunch
				resource["is_public"] = subnet.MapPublicIpOnLaunch != nil && *subnet.MapPublicIpOnLaunch

				// IPv6
				var ipv6Blocks []string
				for _, assoc := range subnet.Ipv6CidrBlockAssociationSet {
					if assoc.Ipv6CidrBlock != nil {
						ipv6Blocks = append(ipv6Blocks, aws.ToString(assoc.Ipv6CidrBlock))
					}
				}
				resource["ipv6_cidr_blocks"] = ipv6Blocks
				resource["has_ipv6"] = len(ipv6Blocks) > 0
				resource["assign_ipv6_on_creation"] = subnet.AssignIpv6AddressOnCreation != nil && *subnet.AssignIpv6AddressOnCreation

				// Tags
				tags := make(map[string]string)
				for _, tag := range subnet.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// VPCEndpointResource fetches VPC endpoints.
type VPCEndpointResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *VPCEndpointResource) Name() string {
	return "aws_vpc_endpoint"
}

// Fetch retrieves all VPC endpoints across configured regions.
func (r *VPCEndpointResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := ec2.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := ec2.NewDescribeVpcEndpointsPaginator(client, &ec2.DescribeVpcEndpointsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, endpoint := range page.VpcEndpoints {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(endpoint.VpcEndpointId)
				resource["region"] = region

				resource["vpc_id"] = aws.ToString(endpoint.VpcId)
				resource["service_name"] = aws.ToString(endpoint.ServiceName)
				resource["state"] = string(endpoint.State)
				resource["endpoint_type"] = string(endpoint.VpcEndpointType)

				// Policy
				if endpoint.PolicyDocument != nil {
					resource["policy_document"] = aws.ToString(endpoint.PolicyDocument)
					resource["has_policy"] = true
				} else {
					resource["has_policy"] = false
				}

				// Private DNS
				resource["private_dns_enabled"] = endpoint.PrivateDnsEnabled != nil && *endpoint.PrivateDnsEnabled

				// Route tables (for Gateway endpoints)
				resource["route_table_ids"] = append([]string{}, endpoint.RouteTableIds...)

				// Subnets (for Interface endpoints)
				resource["subnet_ids"] = append([]string{}, endpoint.SubnetIds...)

				// Security groups (for Interface endpoints)
				sgIDs := make([]string, 0, len(endpoint.Groups))
				for _, sg := range endpoint.Groups {
					sgIDs = append(sgIDs, aws.ToString(sg.GroupId))
				}
				resource["security_group_ids"] = sgIDs

				// DNS entries
				var dnsEntries []string
				for _, dns := range endpoint.DnsEntries {
					if dns.DnsName != nil {
						dnsEntries = append(dnsEntries, aws.ToString(dns.DnsName))
					}
				}
				resource["dns_entries"] = dnsEntries

				// Network interface IDs
				resource["network_interface_ids"] = append([]string{}, endpoint.NetworkInterfaceIds...)

				// Creation time
				if endpoint.CreationTimestamp != nil {
					resource["created_at"] = endpoint.CreationTimestamp.String()
				}

				// Computed
				resource["is_gateway_endpoint"] = endpoint.VpcEndpointType == types.VpcEndpointTypeGateway
				resource["is_interface_endpoint"] = endpoint.VpcEndpointType == types.VpcEndpointTypeInterface
				resource["is_available"] = endpoint.State == types.StateAvailable

				// Tags
				tags := make(map[string]string)
				for _, tag := range endpoint.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// VPCFlowLogResource fetches VPC flow logs.
type VPCFlowLogResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *VPCFlowLogResource) Name() string {
	return "aws_vpc_flowlog"
}

// Fetch retrieves all VPC flow logs across configured regions.
func (r *VPCFlowLogResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := ec2.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := ec2.NewDescribeFlowLogsPaginator(client, &ec2.DescribeFlowLogsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, flowLog := range page.FlowLogs {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(flowLog.FlowLogId)
				resource["region"] = region

				resource["resource_id"] = aws.ToString(flowLog.ResourceId)
				resource["traffic_type"] = string(flowLog.TrafficType)
				resource["log_destination_type"] = string(flowLog.LogDestinationType)
				resource["log_destination"] = aws.ToString(flowLog.LogDestination)
				resource["log_group_name"] = aws.ToString(flowLog.LogGroupName)
				resource["log_format"] = aws.ToString(flowLog.LogFormat)
				resource["status"] = aws.ToString(flowLog.FlowLogStatus)

				// IAM role
				resource["deliver_logs_permission_arn"] = aws.ToString(flowLog.DeliverLogsPermissionArn)

				// Max aggregation interval
				resource["max_aggregation_interval"] = flowLog.MaxAggregationInterval

				// Creation time
				if flowLog.CreationTime != nil {
					resource["created_at"] = flowLog.CreationTime.String()
				}

				// Computed
				resource["is_active"] = aws.ToString(flowLog.FlowLogStatus) == "ACTIVE"
				resource["captures_all_traffic"] = flowLog.TrafficType == types.TrafficTypeAll
				resource["captures_accept_traffic"] = flowLog.TrafficType == types.TrafficTypeAccept || flowLog.TrafficType == types.TrafficTypeAll
				resource["captures_reject_traffic"] = flowLog.TrafficType == types.TrafficTypeReject || flowLog.TrafficType == types.TrafficTypeAll

				// Destination type
				resource["logs_to_cloudwatch"] = flowLog.LogDestinationType == types.LogDestinationTypeCloudWatchLogs
				resource["logs_to_s3"] = flowLog.LogDestinationType == types.LogDestinationTypeS3

				// Tags
				tags := make(map[string]string)
				for _, tag := range flowLog.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
