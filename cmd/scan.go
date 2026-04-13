package cmd

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/report"
	"github.com/cloudmechanic-cli/cloudmechanic/internal/scanner"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	flagRegion     string
	flagProfile    string
	flagOutput     string
	flagAllRegions bool
)

func init() {
	scanCmd.Flags().StringVar(&flagRegion, "region", "", "AWS region to scan (defaults to AWS_REGION or config)")
	scanCmd.Flags().StringVar(&flagProfile, "profile", "", "AWS named profile to use")
	scanCmd.Flags().StringVarP(&flagOutput, "output", "o", "table", "Output format: table, json, csv")
	scanCmd.Flags().BoolVar(&flagAllRegions, "all-regions", false, "Scan all available AWS regions")
	rootCmd.AddCommand(scanCmd)
}

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Run all scanners against your AWS account",
	RunE:  runScan,
}

func runScan(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	start := time.Now()

	cfg, regions, err := LoadAWSConfig(ctx, flagRegion, flagProfile, flagAllRegions)
	if err != nil {
		return err
	}

	if flagOutput == "table" {
		info := color.New(color.FgCyan)
		if len(regions) > 1 {
			info.Fprintf(os.Stderr, "Scanning %d regions...\n", len(regions))
		}
	}

	allScanners := BuildScanners(cfg, regions)
	issues, errs := RunScanners(ctx, allScanners)

	report.Print(os.Stdout, issues, errs, time.Since(start), flagOutput)

	return nil
}

// LoadAWSConfig builds the AWS config and resolves target regions.
func LoadAWSConfig(ctx context.Context, region, profile string, allRegions bool) (aws.Config, []string, error) {
	var cfgOpts []func(*awsconfig.LoadOptions) error
	if region != "" {
		cfgOpts = append(cfgOpts, awsconfig.WithRegion(region))
	}
	if profile != "" {
		cfgOpts = append(cfgOpts, awsconfig.WithSharedConfigProfile(profile))
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, cfgOpts...)
	if err != nil {
		return aws.Config{}, nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	regions, err := resolveRegions(ctx, cfg, allRegions)
	if err != nil {
		return aws.Config{}, nil, err
	}

	return cfg, regions, nil
}

// regionScanner wraps a Scanner and tags all issues with the region they came from.
type regionScanner struct {
	inner  scanner.Scanner
	region string
}

func (rs *regionScanner) Name() string { return rs.inner.Name() }

func (rs *regionScanner) Scan(ctx context.Context) ([]scanner.Issue, error) {
	issues, err := rs.inner.Scan(ctx)
	for i := range issues {
		issues[i].Region = rs.region
	}
	return issues, err
}

// BuildScanners creates all scanner instances for the given regions.
func BuildScanners(cfg aws.Config, regions []string) []scanner.Scanner {
	var allScanners []scanner.Scanner
	for _, region := range regions {
		regionCfg := cfg.Copy()
		regionCfg.Region = region

		ec2Client := ec2.NewFromConfig(regionCfg)
		s3Client := s3.NewFromConfig(regionCfg)
		rdsClient := rds.NewFromConfig(regionCfg)
		cwClient := cloudwatch.NewFromConfig(regionCfg)
		stsClient := sts.NewFromConfig(regionCfg)
		ddbClient := dynamodb.NewFromConfig(regionCfg)
		lambdaClient := lambda.NewFromConfig(regionCfg)

		regional := []scanner.Scanner{
			// EC2
			&scanner.UnattachedEBSScanner{Client: ec2Client},
			&scanner.OpenSecurityGroupScanner{Client: ec2Client},
			&scanner.UnusedEIPScanner{Client: ec2Client},
			&scanner.OldSnapshotScanner{EC2: ec2Client, STS: stsClient},
			// VPC
			&scanner.UnusedNATGatewayScanner{Client: ec2Client},
			&scanner.VPCFlowLogsScanner{Client: ec2Client},
			// S3
			&scanner.PublicS3Scanner{Client: s3Client},
			&scanner.S3EncryptionScanner{Client: s3Client},
			&scanner.S3VersioningScanner{Client: s3Client},
			// RDS
			&scanner.IdleRDSScanner{RDS: rdsClient, CloudWatch: cwClient},
			// DynamoDB
			&scanner.DynamoDBBackupScanner{Client: ddbClient},
			&scanner.DynamoDBUnusedScanner{Client: ddbClient},
			// Lambda
			&scanner.LambdaRuntimeScanner{Client: lambdaClient},
			&scanner.LambdaPublicURLScanner{Client: lambdaClient},
		}

		for _, s := range regional {
			allScanners = append(allScanners, &regionScanner{inner: s, region: region})
		}
	}

	// IAM is global — only run once regardless of region count.
	iamClient := iam.NewFromConfig(cfg)
	allScanners = append(allScanners, &regionScanner{
		inner:  &scanner.NoMFAScanner{Client: iamClient},
		region: "global",
	})

	return allScanners
}

// resolveRegions returns the list of regions to scan.
func resolveRegions(ctx context.Context, cfg aws.Config, allRegions bool) ([]string, error) {
	if !allRegions {
		region := cfg.Region
		if region == "" {
			region = "us-east-1"
		}
		return []string{region}, nil
	}

	ec2Client := ec2.NewFromConfig(cfg)
	out, err := ec2Client.DescribeRegions(ctx, &ec2.DescribeRegionsInput{
		Filters: []ec2types.Filter{
			{
				Name:   aws.String("opt-in-status"),
				Values: []string{"opt-in-not-required", "opted-in"},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("unable to list regions: %w", err)
	}

	var regions []string
	for _, r := range out.Regions {
		regions = append(regions, aws.ToString(r.RegionName))
	}
	return regions, nil
}

// RunScanners executes all scanners concurrently using goroutines and channels.
func RunScanners(ctx context.Context, scanners []scanner.Scanner) ([]scanner.Issue, []error) {
	type result struct {
		issues []scanner.Issue
		err    error
	}

	ch := make(chan result, len(scanners))
	var wg sync.WaitGroup

	for _, s := range scanners {
		wg.Add(1)
		go func(s scanner.Scanner) {
			defer wg.Done()
			issues, err := s.Scan(ctx)
			ch <- result{issues: issues, err: err}
		}(s)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	var allIssues []scanner.Issue
	var allErrors []error
	for r := range ch {
		if r.err != nil {
			allErrors = append(allErrors, r.err)
			continue
		}
		allIssues = append(allIssues, r.issues...)
	}

	return allIssues, allErrors
}
