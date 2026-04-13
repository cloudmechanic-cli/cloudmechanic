package scanner

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// DynamoDBAPI is the interface for DynamoDB calls used by scanners.
type DynamoDBAPI interface {
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
	DescribeContinuousBackups(ctx context.Context, params *dynamodb.DescribeContinuousBackupsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeContinuousBackupsOutput, error)
	DescribeTable(ctx context.Context, params *dynamodb.DescribeTableInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DescribeTableOutput, error)
}

// DynamoDBBackupScanner finds DynamoDB tables without Point-in-Time Recovery enabled.
type DynamoDBBackupScanner struct {
	Client DynamoDBAPI
}

func (s *DynamoDBBackupScanner) Name() string {
	return "DynamoDB Without Backups"
}

func (s *DynamoDBBackupScanner) Scan(ctx context.Context) ([]Issue, error) {
	tables, err := s.listAllTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("dynamodb backup scanner: %w", err)
	}

	var issues []Issue
	for _, table := range tables {
		out, err := s.Client.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
			TableName: aws.String(table),
		})
		if err != nil {
			continue
		}

		desc := out.ContinuousBackupsDescription
		if desc == nil || desc.PointInTimeRecoveryDescription == nil ||
			desc.PointInTimeRecoveryDescription.PointInTimeRecoveryStatus != ddbtypes.PointInTimeRecoveryStatusEnabled {
			issues = append(issues, Issue{
				Severity:    SeverityCritical,
				Scanner:     s.Name(),
				ResourceID:  table,
				Description: fmt.Sprintf("DynamoDB table %s does not have Point-in-Time Recovery enabled", table),
				Suggestion:  "Enable PITR to protect against accidental deletes and allow recovery to any point in the last 35 days.",
			})
		}
	}

	return issues, nil
}

// DynamoDBUnusedScanner finds DynamoDB tables with provisioned capacity and zero consumed read/write.
type DynamoDBUnusedScanner struct {
	Client DynamoDBAPI
}

func (s *DynamoDBUnusedScanner) Name() string {
	return "DynamoDB Provisioned Capacity"
}

func (s *DynamoDBUnusedScanner) Scan(ctx context.Context) ([]Issue, error) {
	tables, err := s.listAllTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("dynamodb unused scanner: %w", err)
	}

	var issues []Issue
	for _, table := range tables {
		out, err := s.Client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(table),
		})
		if err != nil {
			continue
		}

		desc := out.Table
		if desc == nil || desc.BillingModeSummary == nil {
			// No billing mode summary means provisioned (the default).
			if desc != nil && desc.ProvisionedThroughput != nil {
				rcu := aws.ToInt64(desc.ProvisionedThroughput.ReadCapacityUnits)
				wcu := aws.ToInt64(desc.ProvisionedThroughput.WriteCapacityUnits)
				if rcu > 0 || wcu > 0 {
					issues = append(issues, Issue{
						Severity:    SeverityWarning,
						Scanner:     s.Name(),
						ResourceID:  table,
						Description: fmt.Sprintf("DynamoDB table %s uses provisioned capacity (RCU: %d, WCU: %d) — consider on-demand if traffic is unpredictable", table, rcu, wcu),
						Suggestion:  "Switch to on-demand (PAY_PER_REQUEST) billing to avoid paying for unused capacity, or enable auto-scaling.",
					})
				}
			}
			continue
		}

		if desc.BillingModeSummary.BillingMode == ddbtypes.BillingModeProvisioned {
			rcu := aws.ToInt64(desc.ProvisionedThroughput.ReadCapacityUnits)
			wcu := aws.ToInt64(desc.ProvisionedThroughput.WriteCapacityUnits)
			issues = append(issues, Issue{
				Severity:    SeverityWarning,
				Scanner:     s.Name(),
				ResourceID:  table,
				Description: fmt.Sprintf("DynamoDB table %s uses provisioned capacity (RCU: %d, WCU: %d) — consider on-demand if traffic is unpredictable", table, rcu, wcu),
				Suggestion:  "Switch to on-demand (PAY_PER_REQUEST) billing to avoid paying for unused capacity, or enable auto-scaling.",
			})
		}
	}

	return issues, nil
}

func (s *DynamoDBBackupScanner) listAllTables(ctx context.Context) ([]string, error) {
	return listDynamoDBTables(ctx, s.Client)
}

func (s *DynamoDBUnusedScanner) listAllTables(ctx context.Context) ([]string, error) {
	return listDynamoDBTables(ctx, s.Client)
}

func listDynamoDBTables(ctx context.Context, client DynamoDBAPI) ([]string, error) {
	var tables []string
	var lastEval *string

	for {
		out, err := client.ListTables(ctx, &dynamodb.ListTablesInput{
			ExclusiveStartTableName: lastEval,
		})
		if err != nil {
			return nil, err
		}
		tables = append(tables, out.TableNames...)
		if out.LastEvaluatedTableName == nil {
			break
		}
		lastEval = out.LastEvaluatedTableName
	}

	return tables, nil
}
