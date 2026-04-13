package scanner

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// LambdaAPI is the interface for Lambda calls used by scanners.
type LambdaAPI interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
}

// deprecatedRuntimes maps runtime identifiers to their EOL status.
var deprecatedRuntimes = map[string]string{
	"python2.7":    "EOL since July 2022",
	"python3.6":    "EOL since July 2022",
	"python3.7":    "EOL since December 2023",
	"python3.8":    "EOL since October 2024",
	"nodejs10.x":   "EOL since July 2022",
	"nodejs12.x":   "EOL since March 2023",
	"nodejs14.x":   "EOL since December 2023",
	"nodejs16.x":   "EOL since June 2024",
	"dotnetcore2.1": "EOL since January 2022",
	"dotnetcore3.1": "EOL since April 2023",
	"dotnet5.0":    "EOL since February 2024",
	"ruby2.7":      "EOL since December 2023",
	"java8":        "EOL since January 2024",
	"go1.x":        "EOL since January 2024",
}

// LambdaRuntimeScanner finds Lambda functions running on deprecated runtimes.
type LambdaRuntimeScanner struct {
	Client LambdaAPI
}

func (s *LambdaRuntimeScanner) Name() string {
	return "Lambda Deprecated Runtimes"
}

func (s *LambdaRuntimeScanner) Scan(ctx context.Context) ([]Issue, error) {
	functions, err := s.listAllFunctions(ctx)
	if err != nil {
		return nil, fmt.Errorf("lambda runtime scanner: %w", err)
	}

	var issues []Issue
	for _, fn := range functions {
		fnName := aws.ToString(fn.FunctionName)
		runtime := string(fn.Runtime)

		if runtime == "" {
			continue // container image or custom runtime
		}

		eolInfo, deprecated := deprecatedRuntimes[runtime]
		if deprecated {
			issues = append(issues, Issue{
				Severity:    SeverityCritical,
				Scanner:     s.Name(),
				ResourceID:  fnName,
				Description: fmt.Sprintf("Lambda function %s uses deprecated runtime %s (%s)", fnName, runtime, eolInfo),
				Suggestion:  "Upgrade to a supported runtime to continue receiving security patches and AWS support.",
			})
		}
	}

	return issues, nil
}

// LambdaPublicURLScanner finds Lambda functions with public function URLs (no auth).
type LambdaPublicURLScanner struct {
	Client LambdaListFunctionURLAPI
}

// LambdaListFunctionURLAPI extends the base Lambda interface with URL config.
type LambdaListFunctionURLAPI interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
	GetFunctionUrlConfig(ctx context.Context, params *lambda.GetFunctionUrlConfigInput, optFns ...func(*lambda.Options)) (*lambda.GetFunctionUrlConfigOutput, error)
}

func (s *LambdaPublicURLScanner) Name() string {
	return "Lambda Public Function URLs"
}

func (s *LambdaPublicURLScanner) Scan(ctx context.Context) ([]Issue, error) {
	functions, err := listAllLambdaFunctions(ctx, s.Client)
	if err != nil {
		return nil, fmt.Errorf("lambda url scanner: %w", err)
	}

	var issues []Issue
	for _, fn := range functions {
		fnName := aws.ToString(fn.FunctionName)

		urlCfg, err := s.Client.GetFunctionUrlConfig(ctx, &lambda.GetFunctionUrlConfigInput{
			FunctionName: fn.FunctionName,
		})
		if err != nil {
			// No function URL configured — that's fine.
			if strings.Contains(err.Error(), "ResourceNotFoundException") {
				continue
			}
			continue
		}

		if urlCfg.AuthType == lambdatypes.FunctionUrlAuthTypeNone {
			issues = append(issues, Issue{
				Severity:    SeverityCritical,
				Scanner:     s.Name(),
				ResourceID:  fnName,
				Description: fmt.Sprintf("Lambda function %s has a public Function URL with no authentication", fnName),
				Suggestion:  "Set AuthType to AWS_IAM or add a CloudFront distribution with authorization in front of the URL.",
			})
		}
	}

	return issues, nil
}

func (s *LambdaRuntimeScanner) listAllFunctions(ctx context.Context) ([]lambdatypes.FunctionConfiguration, error) {
	return listAllLambdaFunctions(ctx, s.Client)
}

func listAllLambdaFunctions(ctx context.Context, client LambdaAPI) ([]lambdatypes.FunctionConfiguration, error) {
	var functions []lambdatypes.FunctionConfiguration
	var marker *string

	for {
		out, err := client.ListFunctions(ctx, &lambda.ListFunctionsInput{
			Marker: marker,
		})
		if err != nil {
			return nil, err
		}
		functions = append(functions, out.Functions...)
		if out.NextMarker == nil {
			break
		}
		marker = out.NextMarker
	}

	return functions, nil
}
