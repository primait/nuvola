package sts

import (
	"context"
	"errors"

	nuvolaerror "github.com/primait/nuvola/tools/error"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// aws sts get-caller-identity
func Whoami(cfg aws.Config) *sts.GetCallerIdentityOutput {
	var re *http.ResponseError

	output, err := sts.NewFromConfig(cfg).GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "STS", "GetCallerIdentity")
	}

	return output
}
