package sts

import (
	"context"
	"errors"
	"log"

	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// aws sts get-caller-identity
func Whoami(cfg awsconfig.AWSConfig) *sts.GetCallerIdentityOutput {
	var re *http.ResponseError

	output, err := sts.NewFromConfig(cfg.Config).GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if errors.As(err, &re) {
		log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}

	return output
}
