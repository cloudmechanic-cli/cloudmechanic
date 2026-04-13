package scanner

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type mockVpcAPI struct {
	natGateways *ec2.DescribeNatGatewaysOutput
	vpcs        *ec2.DescribeVpcsOutput
	flowLogs    *ec2.DescribeFlowLogsOutput
	routeTables *ec2.DescribeRouteTablesOutput
}

func (m *mockVpcAPI) DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error) {
	return m.natGateways, nil
}

func (m *mockVpcAPI) DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
	return m.vpcs, nil
}

func (m *mockVpcAPI) DescribeFlowLogs(ctx context.Context, params *ec2.DescribeFlowLogsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeFlowLogsOutput, error) {
	return m.flowLogs, nil
}

func (m *mockVpcAPI) DescribeRouteTables(ctx context.Context, params *ec2.DescribeRouteTablesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeRouteTablesOutput, error) {
	return m.routeTables, nil
}

func TestUnusedNATGatewayScanner_FindsUnused(t *testing.T) {
	s := &UnusedNATGatewayScanner{
		Client: &mockVpcAPI{
			natGateways: &ec2.DescribeNatGatewaysOutput{
				NatGateways: []types.NatGateway{
					{
						NatGatewayId: aws.String("nat-unused"),
						VpcId:        aws.String("vpc-123"),
						State:        types.NatGatewayStateAvailable,
					},
				},
			},
			routeTables: &ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{
					{
						Routes: []types.Route{
							{DestinationCidrBlock: aws.String("0.0.0.0/0"), GatewayId: aws.String("igw-abc")},
						},
					},
				},
			},
		},
	}

	issues, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != SeverityWarning {
		t.Errorf("expected WARNING, got %s", issues[0].Severity)
	}
}

func TestUnusedNATGatewayScanner_UsedInRoute(t *testing.T) {
	s := &UnusedNATGatewayScanner{
		Client: &mockVpcAPI{
			natGateways: &ec2.DescribeNatGatewaysOutput{
				NatGateways: []types.NatGateway{
					{
						NatGatewayId: aws.String("nat-used"),
						VpcId:        aws.String("vpc-123"),
						State:        types.NatGatewayStateAvailable,
					},
				},
			},
			routeTables: &ec2.DescribeRouteTablesOutput{
				RouteTables: []types.RouteTable{
					{
						Routes: []types.Route{
							{DestinationCidrBlock: aws.String("0.0.0.0/0"), NatGatewayId: aws.String("nat-used")},
						},
					},
				},
			},
		},
	}

	issues, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected 0 issues, got %d", len(issues))
	}
}

func TestVPCFlowLogsScanner_NoFlowLogs(t *testing.T) {
	s := &VPCFlowLogsScanner{
		Client: &mockVpcAPI{
			vpcs: &ec2.DescribeVpcsOutput{
				Vpcs: []types.Vpc{
					{VpcId: aws.String("vpc-nofl")},
				},
			},
			flowLogs: &ec2.DescribeFlowLogsOutput{FlowLogs: []types.FlowLog{}},
		},
	}

	issues, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}
	if issues[0].Severity != SeverityCritical {
		t.Errorf("expected CRITICAL, got %s", issues[0].Severity)
	}
}

func TestVPCFlowLogsScanner_HasFlowLogs(t *testing.T) {
	s := &VPCFlowLogsScanner{
		Client: &mockVpcAPI{
			vpcs: &ec2.DescribeVpcsOutput{
				Vpcs: []types.Vpc{
					{VpcId: aws.String("vpc-good")},
				},
			},
			flowLogs: &ec2.DescribeFlowLogsOutput{
				FlowLogs: []types.FlowLog{
					{ResourceId: aws.String("vpc-good")},
				},
			},
		},
	}

	issues, err := s.Scan(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(issues) != 0 {
		t.Fatalf("expected 0 issues, got %d", len(issues))
	}
}
