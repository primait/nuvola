package awsconnector

import (
	"context"
	"os"

	"github.com/primait/nuvola/pkg/connector/services/aws/database"
	"github.com/primait/nuvola/pkg/connector/services/aws/ec2"
	"github.com/primait/nuvola/pkg/connector/services/aws/iam"
	"github.com/primait/nuvola/pkg/connector/services/aws/lambda"
	"github.com/primait/nuvola/pkg/connector/services/aws/s3"
	"github.com/primait/nuvola/pkg/connector/services/aws/sts"
	"github.com/primait/nuvola/pkg/io/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	ActionsMap   map[string][]string
	ActionsList  []string // len(unique(ActionList)) ~= 13k
	Conditions   map[string]string
	countRetries = 100
)

func InitAWSConfiguration(profile string, awsEndpoint string) (awsc AWSConfig) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if awsEndpoint != "" {
			return aws.Endpoint{
				PartitionID:   "aws",
				URL:           awsEndpoint,
				SigningRegion: os.Getenv("AWS_DEFAULT_REGION"),
			}, nil
		}

		// returning EndpointNotFoundError will allow the service to fallback to it's default resolution
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), countRetries)
		}),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	cfg.RetryMode = aws.RetryModeStandard
	awsc = AWSConfig{Profile: profile, Config: cfg}
	SetActions()
	// Get the AWS regions dynamically
	ec2.ListAndSaveRegions(cfg)
	iam.ActionsList = unique(ActionsList)
	iam.ActionsMap = ActionsMap
	return
}

func (ac *AWSConfig) TestConnection() bool {
	_, err := ac.Credentials.Retrieve(context.TODO())
	return err == nil
}

func (ac *AWSConfig) DumpWhoami() interface{} {
	return sts.Whoami(ac.Config)
}

func (ac *AWSConfig) DumpCredentialReport() interface{} {
	report := iam.GetCredentialReport(ac.Config)
	return report
}

func (ac *AWSConfig) DumpIAMGroups() interface{} {
	groups := iam.ListGroups(ac.Config)
	return groups
}

func (ac *AWSConfig) DumpIAMUsers() interface{} {
	report := iam.GetCredentialReport(ac.Config)
	users := iam.ListUsers(ac.Config, report)
	return users
}

func (ac *AWSConfig) DumpIAMRoles() interface{} {
	return iam.ListRoles(ac.Config)
}

func (ac *AWSConfig) DumpBuckets() interface{} {
	buckets := s3.ListBuckets(ac.Config)
	return buckets
}

func (ac *AWSConfig) DumpEC2Instances() interface{} {
	ec2s, err := ec2.ListInstances(ac.Config)
	if err != nil {
		logging.HandleError(err, "EC2", "")
	}
	return ec2s
}

func (ac *AWSConfig) DumpVpcs() interface{} {
	return ec2.ListVpcs(ac.Config)
}

func (ac *AWSConfig) DumpLambdas() interface{} {
	return lambda.ListFunctions(ac.Config)
}

func (ac *AWSConfig) DumpRDS() interface{} {
	rds, err := database.ListRDS(ac.Config)
	if err != nil {
		logging.HandleError(err, "RDS", "")
	}
	return rds
}

func (ac *AWSConfig) DumpDynamoDBs() interface{} {
	tables := database.ListDynamoDBs(ac.Config)
	return tables
}

func (ac *AWSConfig) DumpRedshiftDBs() interface{} {
	redshift := database.ListRedshiftDBs(ac.Config)
	return redshift
}

func (ac *AWSConfig) DumpAll() map[string]interface{} {
	// The order is important!
	return map[string]interface{}{
		"Whoami":           ac.DumpWhoami(),
		"CredentialReport": ac.DumpCredentialReport(),
		"Groups":           ac.DumpIAMGroups(),
		"Users":            ac.DumpIAMUsers(),
		"Roles":            ac.DumpIAMRoles(),
		"Buckets":          ac.DumpBuckets(),
		"EC2s":             ac.DumpEC2Instances(),
		"VPCs":             ac.DumpVpcs(),
		"Lambdas":          ac.DumpLambdas(),
		"RDS":              ac.DumpRDS(),
		"DynamoDBs":        ac.DumpDynamoDBs(),
		"RedshiftDBs":      ac.DumpRedshiftDBs(),
	}
}
