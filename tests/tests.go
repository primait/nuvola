package tests

import (
	"context"

	awsconnector "nuvola/connector/services/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
)

func InitAWSConfiguration(profile string) (awsc awsconnector.AWSConfig) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile), config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), 20)
	}))
	cfg.RetryMode = aws.RetryModeStandard
	awsc = awsconnector.AWSConfig{Profile: profile, Config: cfg}
	return
}
