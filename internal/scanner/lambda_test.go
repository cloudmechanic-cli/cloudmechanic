package scanner

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type mockLambda struct {
	functions *lambda.ListFunctionsOutput
}

func (m *mockLambda) ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
	return m.functions, nil
}

func TestLambdaRuntimeScanner_FindsDeprecated(t *testing.T) {
	s := &LambdaRuntimeScanner{
		Client: &mockLambda{
			functions: &lambda.ListFunctionsOutput{
				Functions: []lambdatypes.FunctionConfiguration{
					{FunctionName: aws.String("old-func"), Runtime: lambdatypes.RuntimePython27},
					{FunctionName: aws.String("new-func"), Runtime: lambdatypes.RuntimePython312},
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
	if issues[0].ResourceID != "old-func" {
		t.Errorf("expected old-func, got %s", issues[0].ResourceID)
	}
	if issues[0].Severity != SeverityCritical {
		t.Errorf("expected CRITICAL, got %s", issues[0].Severity)
	}
}

func TestLambdaRuntimeScanner_AllCurrent(t *testing.T) {
	s := &LambdaRuntimeScanner{
		Client: &mockLambda{
			functions: &lambda.ListFunctionsOutput{
				Functions: []lambdatypes.FunctionConfiguration{
					{FunctionName: aws.String("fn1"), Runtime: lambdatypes.RuntimePython312},
					{FunctionName: aws.String("fn2"), Runtime: lambdatypes.RuntimeNodejs20x},
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
