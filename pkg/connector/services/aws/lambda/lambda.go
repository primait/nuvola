package lambda

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/primait/nuvola/pkg/connector/services/aws/ec2"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/sourcegraph/conc/iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// aws iam list-users
func ListFunctions(cfg aws.Config) (lambdas []*Lambda) {
	var lambdaClient = LambdaClient{Config: cfg}

	for i := range ec2.Regions {
		cfg.Region = ec2.Regions[i]
		lambdaClient.client = lambda.NewFromConfig(cfg)
		lambdas = append(lambdas, lambdaClient.listFunctionsForRegion()...)
	}

	return
}

func (lc *LambdaClient) listFunctionsForRegion() (lambdas []*Lambda) {
	output, err := lc.client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})

	if errors.As(err, &re) {
		logging.HandleAWSError(re, "Lambda", "ListFunctions")
	}

	lambdas = iter.Map(output.Functions, func(lambda *types.FunctionConfiguration) *Lambda {
		return &Lambda{
			FunctionConfiguration: *lambda,
			FunctionCodeLocation:  lc.getFunctionCodeLocation(aws.ToString(lambda.FunctionName)),
			Policy:                lc.getPolicy(aws.ToString(lambda.FunctionName)),
		}
	})
	return
}

func (lc *LambdaClient) getFunctionCodeLocation(name string) types.FunctionCodeLocation {
	output, err := lc.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
		FunctionName: &name,
	})
	if errors.As(err, &re) {
		logging.HandleAWSError(re, "Lambda", "GetFunction")
	}

	return *output.Code
}

func (lc *LambdaClient) getPolicy(name string) (policyDocument lambdaPolicyDocument) {
	var re *http.ResponseError

	output, err := lc.client.GetPolicy(context.TODO(), &lambda.GetPolicyInput{
		FunctionName: &name,
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() != 404 { // Function can't have a policy
			logging.HandleAWSError(re, "Lambda", "GetPolicy")
		}
		return policyDocument
	}

	if output.Policy != nil {
		err := json.Unmarshal([]byte(aws.ToString(output.Policy)), &policyDocument)
		if err != nil {
			logging.HandleError(err, "Lambda", "Umarshalling policyDocument")
		}
	}

	return policyDocument
}
