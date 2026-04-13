package scanner

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type mockDynamoDB struct {
	tables    *dynamodb.ListTablesOutput
	backups   map[string]*dynamodb.DescribeContinuousBackupsOutput
	tableDesc map[string]*dynamodb.DescribeTableOutput
}

func (m *mockDynamoDB) ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
	return m.tables, nil
}

func (m *mockDynamoDB) DescribeContinuousBackups(ctx context.Context, params *dynamodb.DescribeContinuousBackupsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeContinuousBackupsOutput, error) {
	return m.backups[aws.ToString(params.TableName)], nil
}

func (m *mockDynamoDB) DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error) {
	return m.tableDesc[aws.ToString(params.TableName)], nil
}

func TestDynamoDBBackupScanner_NoPITR(t *testing.T) {
	s := &DynamoDBBackupScanner{
		Client: &mockDynamoDB{
			tables: &dynamodb.ListTablesOutput{TableNames: []string{"users"}},
			backups: map[string]*dynamodb.DescribeContinuousBackupsOutput{
				"users": {
					ContinuousBackupsDescription: &ddbtypes.ContinuousBackupsDescription{
						PointInTimeRecoveryDescription: &ddbtypes.PointInTimeRecoveryDescription{
							PointInTimeRecoveryStatus: ddbtypes.PointInTimeRecoveryStatusDisabled,
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
	if issues[0].Severity != SeverityCritical {
		t.Errorf("expected CRITICAL, got %s", issues[0].Severity)
	}
}

func TestDynamoDBBackupScanner_PITREnabled(t *testing.T) {
	s := &DynamoDBBackupScanner{
		Client: &mockDynamoDB{
			tables: &dynamodb.ListTablesOutput{TableNames: []string{"orders"}},
			backups: map[string]*dynamodb.DescribeContinuousBackupsOutput{
				"orders": {
					ContinuousBackupsDescription: &ddbtypes.ContinuousBackupsDescription{
						PointInTimeRecoveryDescription: &ddbtypes.PointInTimeRecoveryDescription{
							PointInTimeRecoveryStatus: ddbtypes.PointInTimeRecoveryStatusEnabled,
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
