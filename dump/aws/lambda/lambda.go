package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"golang.org/x/sync/semaphore"
)

// aws iam list-users
func ListFunctions(cfg awsconfig.AWSConfig) (Lambdas []*Lambda) {
	var lambdaClient = LambdaClient{Config: cfg}

	for i := range awsconfig.Regions {
		cfg.Config.Region = awsconfig.Regions[i]
		lambdaClient.client = lambda.NewFromConfig(cfg.Config)
		Lambdas = append(Lambdas, lambdaClient.listFunctionsForRegion()...)
	}

	return
}

func (lc *LambdaClient) listFunctionsForRegion() (Lambdas []*Lambda) {
	var (
		re  *awshttp.ResponseError
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(20)) // TODO: parametric
		wg  sync.WaitGroup
	)

	output, err := lc.client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{})

	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			lc.Config.WaitAPILimit()
			return lc.listFunctionsForRegion()
		} else if re.HTTPStatusCode() == 403 {
			return
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "Lambda", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	for _, lambda := range output.Functions {
		wg.Add(1)
		go func(lambda types.FunctionConfiguration) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v\n", err)
			}
			defer sem.Release(1)
			defer wg.Done()
			defer mu.Unlock()
			var item *Lambda = &Lambda{
				FunctionConfiguration: lambda,
				FunctionCodeLocation:  lc.getFunctionCodeLocation(aws.ToString(lambda.FunctionName)),
				Policy:                lc.getPolicy(aws.ToString(lambda.FunctionName)),
			}
			mu.Lock()
			Lambdas = append(Lambdas, item)
		}(lambda)
	}
	wg.Wait()
	return
}

func (lc *LambdaClient) getFunctionCodeLocation(name string) types.FunctionCodeLocation {
	var re *awshttp.ResponseError

	output, err := lc.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{
		FunctionName: &name,
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			lc.Config.WaitAPILimit()
			return lc.getFunctionCodeLocation(name)
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "Lambda", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}
	return *output.Code
}

func (lc *LambdaClient) getPolicy(name string) (policyDocument lambdaPolicyDocument) {
	var re *awshttp.ResponseError

	output, err := lc.client.GetPolicy(context.TODO(), &lambda.GetPolicyInput{
		FunctionName: &name,
	})
	if errors.As(err, &re) {
		switch re.HTTPStatusCode() {
		case 504:
			lc.Config.WaitAPILimit()
			return lc.getPolicy(name)
		case 404:
			return policyDocument
		default:
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "Lambda", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}
	if output.Policy != nil {
		err := json.Unmarshal([]byte(aws.ToString(output.Policy)), &policyDocument)
		if err != nil {
			log.Fatalln(err)
		}
	}

	return policyDocument
}
