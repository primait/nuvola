package connector

import (
	"errors"
	awsconfig "nuvola/connector/services/aws"
	"strings"
)

func NewCloudConnector(profile string, endpointUrl string) (*CloudConnector, error) {
	cc := &CloudConnector{
		AWSConfig: awsconfig.InitAWSConfiguration(profile, endpointUrl),
	}
	if !cc.testConnection("aws") {
		return nil, errors.New("invalid credentials or expired session")
	}
	return cc, nil
}

func SetActions() {
	awsconfig.SetActions()
}

func (cc *CloudConnector) DumpAll(cloudprovider string, c chan map[string]interface{}) {
	switch strings.ToLower(cloudprovider) {
	case "aws":
		whoami := cc.AWSConfig.DumpWhoami()
		c <- map[string]interface{}{
			"Whoami": whoami,
		}
		credentialReport := cc.AWSConfig.DumpCredentialReport()
		c <- map[string]interface{}{
			"CredentialReport": credentialReport,
		}
		groups := cc.AWSConfig.DumpIAMGroups()
		c <- map[string]interface{}{
			"Groups": groups,
		}
		users := cc.AWSConfig.DumpIAMUsers()
		c <- map[string]interface{}{
			"Users": users,
		}
		roles := cc.AWSConfig.DumpIAMRoles()
		c <- map[string]interface{}{
			"Roles": roles,
		}
		buckets := cc.AWSConfig.DumpBuckets()
		c <- map[string]interface{}{
			"Buckets": buckets,
		}
		ec2 := cc.AWSConfig.DumpEC2Instances()
		c <- map[string]interface{}{
			"EC2s": ec2,
		}
		vpc := cc.AWSConfig.DumpVpcs()
		c <- map[string]interface{}{
			"VPCs": vpc,
		}
		lambda := cc.AWSConfig.DumpLambdas()
		c <- map[string]interface{}{
			"Lambdas": lambda,
		}
		rds := cc.AWSConfig.DumpRDS()
		c <- map[string]interface{}{
			"RDS": rds,
		}
		dynamodb := cc.AWSConfig.DumpDynamoDBs()
		c <- map[string]interface{}{
			"DynamoDBs": dynamodb,
		}
		redshift := cc.AWSConfig.DumpRedshiftDBs()
		c <- map[string]interface{}{
			"RedshiftDBs": redshift,
		}
		close(c)
	default:
	}
}

func (cc *CloudConnector) testConnection(cloudprovider string) bool {
	switch strings.ToLower(cloudprovider) {
	case "aws":
		return cc.AWSConfig.TestConnection()
	default:
		return false
	}
}
