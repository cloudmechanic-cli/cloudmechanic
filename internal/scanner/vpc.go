package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2VpcAPI is the interface for the EC2 VPC-related calls.
type EC2VpcAPI interface {
	DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeFlowLogs(ctx context.Context, params *ec2.DescribeFlowLogsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeFlowLogsOutput, error)
	DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error)
}

// UnusedNATGatewayScanner finds NAT Gateways that are not referenced in any route table.
type UnusedNATGatewayScanner struct {
	Client EC2VpcAPI
}

func (s *UnusedNATGatewayScanner) Name() string {
	return "Unused NAT Gateways"
}

func (s *UnusedNATGatewayScanner) Scan(ctx context.Context) ([]Issue, error) {
	natOut, err := s.Client.DescribeNatGateways(ctx, &ec2.DescribeNatGatewaysInput{})
	if err != nil {
		return nil, fmt.Errorf("nat gateway scanner: %w", err)
	}

	// Get all route tables to check which NAT gateways are actually used.
	rtOut, err := s.Client.DescribeRouteTables(ctx, &ec2.DescribeRouteTablesInput{})
	if err != nil {
		return nil, fmt.Errorf("nat gateway scanner: %w", err)
	}

	usedNATs := make(map[string]bool)
	for _, rt := range rtOut.RouteTables {
		for _, route := range rt.Routes {
			if route.NatGatewayId != nil {
				usedNATs[aws.ToString(route.NatGatewayId)] = true
			}
		}
	}

	var issues []Issue
	for _, nat := range natOut.NatGateways {
		if nat.State != types.NatGatewayStateAvailable {
			continue
		}

		natID := aws.ToString(nat.NatGatewayId)
		vpcID := aws.ToString(nat.VpcId)

		if !usedNATs[natID] {
			issues = append(issues, Issue{
				Severity:    SeverityWarning,
				Scanner:     s.Name(),
				ResourceID:  natID,
				Description: fmt.Sprintf("NAT Gateway %s in VPC %s is not referenced in any route table", natID, vpcID),
				Suggestion:  "Delete this NAT Gateway to stop incurring charges (~$32/month + data processing fees).",
			})
		}
	}

	return issues, nil
}

// VPCFlowLogsScanner finds VPCs that do not have Flow Logs enabled.
type VPCFlowLogsScanner struct {
	Client EC2VpcAPI
}

func (s *VPCFlowLogsScanner) Name() string {
	return "VPCs Without Flow Logs"
}

func (s *VPCFlowLogsScanner) Scan(ctx context.Context) ([]Issue, error) {
	vpcOut, err := s.Client.DescribeVpcs(ctx, &ec2.DescribeVpcsInput{})
	if err != nil {
		return nil, fmt.Errorf("vpc flow logs scanner: %w", err)
	}

	flowOut, err := s.Client.DescribeFlowLogs(ctx, &ec2.DescribeFlowLogsInput{})
	if err != nil {
		return nil, fmt.Errorf("vpc flow logs scanner: %w", err)
	}

	// Build a set of VPC IDs that have at least one flow log.
	vpcsWithFlowLogs := make(map[string]bool)
	for _, fl := range flowOut.FlowLogs {
		if aws.ToString(fl.ResourceId) != "" {
			vpcsWithFlowLogs[aws.ToString(fl.ResourceId)] = true
		}
	}

	var issues []Issue
	for _, vpc := range vpcOut.Vpcs {
		vpcID := aws.ToString(vpc.VpcId)

		if vpcsWithFlowLogs[vpcID] {
			continue
		}

		name := vpcID
		for _, tag := range vpc.Tags {
			if aws.ToString(tag.Key) == "Name" {
				name = fmt.Sprintf("%s (%s)", vpcID, aws.ToString(tag.Value))
				break
			}
		}

		issues = append(issues, Issue{
			Severity:    SeverityCritical,
			Scanner:     s.Name(),
			ResourceID:  vpcID,
			Description: fmt.Sprintf("VPC %s does not have Flow Logs enabled", name),
			Suggestion:  "Enable VPC Flow Logs to capture network traffic for security auditing and troubleshooting.",
		})
	}

	return issues, nil
}
