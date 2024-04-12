package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"sync"

	nuvolaerror "github.com/primait/nuvola/tools/error"

	"github.com/primait/nuvola/connector/services/aws/ec2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"golang.org/x/sync/semaphore"
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
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(20))
		wg  sync.WaitGroup
	)

	output, err := lc.client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})

	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "Lambda", "ListFunctions")
	}

	for _, lambda := range output.Functions {
		wg.Add(1)
		go func(lambda types.FunctionConfiguration) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				nuvolaerror.HandleError(err, "Lambda", "listFunctionsForRegion - Acquire Semaphore")
			}
			defer sem.Release(1)
			defer wg.Done()
			defer mu.Unlock()
			var item = &Lambda{
				FunctionConfiguration: lambda,
				FunctionCodeLocation:  lc.getFunctionCodeLocation(aws.ToString(lambda.FunctionName)),
				Policy:                lc.getPolicy(aws.ToString(lambda.FunctionName)),
			}
			mu.Lock()
			lambdas = append(lambdas, item)
		}(lambda)
	}
	wg.Wait()
	return
}

func (lc *LambdaClient) getFunctionCodeLocation(name string) types.FunctionCodeLocation {
	output, err := lc.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
		FunctionName: &name,
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "Lambda", "GetFunction")
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
			nuvolaerror.HandleAWSError(re, "Lambda", "GetPolicy")
		}
		return policyDocument
	}

	if output.Policy != nil {
		err := json.Unmarshal([]byte(aws.ToString(output.Policy)), &policyDocument)
		if err != nil {
			nuvolaerror.HandleError(err, "Lambda", "Umarshalling policyDocument")
		}
	}

	return policyDocument
}
