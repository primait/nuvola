package awsconnector

import (
	"context"

	"nuvola/connector/services/aws/database"
	"nuvola/connector/services/aws/ec2"
	"nuvola/connector/services/aws/iam"
	"nuvola/connector/services/aws/lambda"
	"nuvola/connector/services/aws/s3"
	"nuvola/connector/services/aws/sts"
	nuvolaerror "nuvola/tools/error"

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

func InitAWSConfiguration(profile string) (awsc AWSConfig) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile), config.WithRetryer(func() aws.Retryer {
		return retry.AddWithMaxAttempts(retry.NewStandard(), countRetries)
	}))
	cfg.RetryMode = aws.RetryModeStandard
	awsc = AWSConfig{Profile: profile, Config: cfg}
	SetActions()
	// Get the AWS regions dynamically
	ec2.ListAndSaveRegions(cfg)
	iam.ActionsList = ActionsList
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
		nuvolaerror.HandleError(err, "EC2", "")
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
		nuvolaerror.HandleError(err, "RDS", "")
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
