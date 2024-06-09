package sts

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/primait/nuvola/pkg/io/logging"
)

// aws sts get-caller-identity
func Whoami(cfg aws.Config) *sts.GetCallerIdentityOutput {
	var re *http.ResponseError

	logger := logging.GetLogManager()
	output, err := sts.NewFromConfig(cfg).GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	if errors.As(err, &re) {
		logger.Warn("Error on GetCallerIdentity", "err", re)
	}

	logger.Info("sts get-caller-identity", "account", aws.ToString(output.Account), "arn", aws.ToString(output.Arn))
	return output
}
