package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/kopexa-grc/kspec/core"
)

// EC2SecurityGroupResource fetches EC2 security groups.
type EC2SecurityGroupResource struct {
	conn   *Connection
	client EC2API // optional, for testing
}

// Name returns the resource type name.
func (r *EC2SecurityGroupResource) Name() string {
	return "aws_ec2_securitygroup"
}

// getClient returns the EC2 client for the given region.
func (r *EC2SecurityGroupResource) getClient(region string) EC2API {
	if r.client != nil {
		return r.client
	}
	return ec2.NewFromConfig(r.conn.ConfigForRegion(region))
}

// Fetch retrieves all EC2 security groups across configured regions.
func (r *EC2SecurityGroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := r.getClient(region)

		paginator := ec2.NewDescribeSecurityGroupsPaginator(client, &ec2.DescribeSecurityGroupsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, sg := range page.SecurityGroups {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(sg.GroupId)
				resource["name"] = aws.ToString(sg.GroupName)
				resource["region"] = region
				resource["arn"] = "arn:aws:ec2:" + region + ":" + r.conn.accountID + ":security-group/" + aws.ToString(sg.GroupId)

				resource["description"] = aws.ToString(sg.Description)
				resource["vpc_id"] = aws.ToString(sg.VpcId)
				resource["owner_id"] = aws.ToString(sg.OwnerId)

				// Computed: Is default security group
				resource["is_default"] = aws.ToString(sg.GroupName) == "default"

				// Analyze inbound rules
				var (
					hasUnrestrictedIngress  bool
					allowsSSHFromInternet   bool
					allowsRDPFromInternet   bool
					allowsAllTrafficIngress bool
					inboundRules            []map[string]any
				)

				for _, perm := range sg.IpPermissions {
					rule := map[string]any{
						"protocol":  aws.ToString(perm.IpProtocol),
						"from_port": perm.FromPort,
						"to_port":   perm.ToPort,
					}

					var cidrs []string
					for _, ipRange := range perm.IpRanges {
						cidr := aws.ToString(ipRange.CidrIp)
						cidrs = append(cidrs, cidr)

						// Check for unrestricted access
						if cidr == cidrAllIPv4 || cidr == cidrAllIPv6 {
							hasUnrestrictedIngress = true

							// Check protocol and port
							protocol := aws.ToString(perm.IpProtocol)
							fromPort := perm.FromPort
							toPort := perm.ToPort

							// All traffic (-1 protocol)
							if protocol == "-1" {
								allowsAllTrafficIngress = true
							}

							// SSH (port 22)
							if (protocol == "tcp" || protocol == "-1") &&
								(fromPort == nil || *fromPort <= 22) &&
								(toPort == nil || *toPort >= 22) {
								allowsSSHFromInternet = true
							}

							// RDP (port 3389)
							if (protocol == "tcp" || protocol == "-1") &&
								(fromPort == nil || *fromPort <= 3389) &&
								(toPort == nil || *toPort >= 3389) {
								allowsRDPFromInternet = true
							}
						}
					}

					// IPv6 ranges
					for _, ipv6Range := range perm.Ipv6Ranges {
						cidr := aws.ToString(ipv6Range.CidrIpv6)
						cidrs = append(cidrs, cidr)

						if cidr == cidrAllIPv6 {
							hasUnrestrictedIngress = true
						}
					}

					rule["cidrs"] = cidrs

					// Referenced security groups
					var referencedSGs []string
					for _, refSG := range perm.UserIdGroupPairs {
						referencedSGs = append(referencedSGs, aws.ToString(refSG.GroupId))
					}
					rule["security_groups"] = referencedSGs

					inboundRules = append(inboundRules, rule)
				}

				resource["inbound_rules"] = inboundRules
				resource["inbound_rule_count"] = len(inboundRules)
				resource["has_unrestricted_ingress"] = hasUnrestrictedIngress
				resource["allows_ssh_from_internet"] = allowsSSHFromInternet
				resource["allows_rdp_from_internet"] = allowsRDPFromInternet
				resource["allows_all_traffic_ingress"] = allowsAllTrafficIngress

				// Analyze outbound rules
				var (
					hasUnrestrictedEgress  bool
					allowsAllTrafficEgress bool
					outboundRules          []map[string]any
				)

				for _, perm := range sg.IpPermissionsEgress {
					rule := map[string]any{
						"protocol":  aws.ToString(perm.IpProtocol),
						"from_port": perm.FromPort,
						"to_port":   perm.ToPort,
					}

					var cidrs []string
					for _, ipRange := range perm.IpRanges {
						cidr := aws.ToString(ipRange.CidrIp)
						cidrs = append(cidrs, cidr)

						if cidr == cidrAllIPv4 || cidr == cidrAllIPv6 {
							hasUnrestrictedEgress = true

							if aws.ToString(perm.IpProtocol) == "-1" {
								allowsAllTrafficEgress = true
							}
						}
					}

					for _, ipv6Range := range perm.Ipv6Ranges {
						cidr := aws.ToString(ipv6Range.CidrIpv6)
						cidrs = append(cidrs, cidr)

						if cidr == cidrAllIPv6 {
							hasUnrestrictedEgress = true
						}
					}

					rule["cidrs"] = cidrs

					var referencedSGs []string
					for _, refSG := range perm.UserIdGroupPairs {
						referencedSGs = append(referencedSGs, aws.ToString(refSG.GroupId))
					}
					rule["security_groups"] = referencedSGs

					outboundRules = append(outboundRules, rule)
				}

				resource["outbound_rules"] = outboundRules
				resource["outbound_rule_count"] = len(outboundRules)
				resource["has_unrestricted_egress"] = hasUnrestrictedEgress
				resource["allows_all_traffic_egress"] = allowsAllTrafficEgress

				// Computed: Allows all traffic (both directions unrestricted)
				resource["allows_all_traffic"] = allowsAllTrafficIngress && allowsAllTrafficEgress

				// Check if security group is attached to any resources
				// This requires checking network interfaces
				niResp, err := client.DescribeNetworkInterfaces(ctx, &ec2.DescribeNetworkInterfacesInput{
					Filters: []types.Filter{
						{
							Name:   aws.String("group-id"),
							Values: []string{aws.ToString(sg.GroupId)},
						},
					},
				})
				if err == nil {
					resource["is_attached"] = len(niResp.NetworkInterfaces) > 0
					resource["attached_resource_count"] = len(niResp.NetworkInterfaces)
				} else {
					resource["is_attached"] = false
					resource["attached_resource_count"] = 0
				}

				// Tags
				tags := make(map[string]string)
				for _, tag := range sg.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// EC2KeyPairResource fetches EC2 key pairs.
type EC2KeyPairResource struct {
	conn   *Connection
	client EC2API // optional, for testing
}

// Name returns the resource type name.
func (r *EC2KeyPairResource) Name() string {
	return "aws_ec2_keypair"
}

// getClient returns the EC2 client for the given region.
func (r *EC2KeyPairResource) getClient(region string) EC2API {
	if r.client != nil {
		return r.client
	}
	return ec2.NewFromConfig(r.conn.ConfigForRegion(region))
}

// Fetch retrieves all EC2 key pairs across configured regions.
func (r *EC2KeyPairResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := r.getClient(region)

		resp, err := client.DescribeKeyPairs(ctx, &ec2.DescribeKeyPairsInput{
			IncludePublicKey: aws.Bool(true),
		})
		if err != nil {
			continue
		}

		for _, keyPair := range resp.KeyPairs {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(keyPair.KeyPairId)
			resource["name"] = aws.ToString(keyPair.KeyName)
			resource["region"] = region
			resource["arn"] = "arn:aws:ec2:" + region + ":" + r.conn.accountID + ":key-pair/" + aws.ToString(keyPair.KeyPairId)

			resource["key_fingerprint"] = aws.ToString(keyPair.KeyFingerprint)
			resource["key_type"] = string(keyPair.KeyType)

			// Public key (if available)
			if keyPair.PublicKey != nil {
				resource["public_key"] = aws.ToString(keyPair.PublicKey)
			}

			// Create time
			if keyPair.CreateTime != nil {
				resource["created_at"] = keyPair.CreateTime.String()
			}

			// Computed: Key type analysis
			keyType := string(keyPair.KeyType)
			resource["is_rsa"] = keyType == "rsa"
			resource["is_ed25519"] = keyType == "ed25519"

			// RSA keys are considered less secure than ed25519
			resource["is_modern_key_type"] = keyType == "ed25519"

			// Tags
			tags := make(map[string]string)
			for _, tag := range keyPair.Tags {
				tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
			}
			resource["tags"] = tags

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
