package connector

import (
	"errors"
	"strings"

	awsconfig "github.com/primait/nuvola/pkg/connector/services/aws"
	"github.com/primait/nuvola/pkg/io/logging"
)

func NewCloudConnector(profile string, endpointUrl string) (*CloudConnector, error) {
	cc := &CloudConnector{
		AWSConfig: awsconfig.InitAWSConfiguration(profile, endpointUrl),
		logger:    logging.GetLogManager(),
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
		dumpFunctions := []struct {
			name string
			dump func() interface{}
		}{
			{"Whoami", cc.AWSConfig.DumpWhoami},
			{"CredentialReport", cc.AWSConfig.DumpCredentialReport},
			{"Groups", cc.AWSConfig.DumpIAMGroups},
			{"Users", cc.AWSConfig.DumpIAMUsers},
			{"Roles", cc.AWSConfig.DumpIAMRoles},
			{"Buckets", cc.AWSConfig.DumpBuckets},
			{"EC2s", cc.AWSConfig.DumpEC2Instances},
			{"VPCs", cc.AWSConfig.DumpVpcs},
			{"Lambdas", cc.AWSConfig.DumpLambdas},
			{"RDS", cc.AWSConfig.DumpRDS},
			{"DynamoDBs", cc.AWSConfig.DumpDynamoDBs},
			{"RedshiftDBs", cc.AWSConfig.DumpRedshiftDBs},
		}

		for _, df := range dumpFunctions {
			data := df.dump()
			if data != nil {
				c <- map[string]interface{}{
					df.name: data,
				}
			}
		}
	default:
		cc.logger.Error("unsupported cloud provider", "cloudprovider", cloudprovider)
	}
}

func (cc *CloudConnector) testConnection(cloudprovider string) bool {
	switch strings.ToLower(cloudprovider) {
	case "aws":
		return cc.AWSConfig.TestConnection()
	default:
		cc.logger.Error("Unsupported cloud provider", "cloudprovider", cloudprovider)
		return false
	}
}
