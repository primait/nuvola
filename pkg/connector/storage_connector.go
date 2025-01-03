package connector

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/primait/nuvola/pkg/connector/services/aws/database"
	"github.com/primait/nuvola/pkg/connector/services/aws/ec2"
	"github.com/primait/nuvola/pkg/connector/services/aws/iam"
	"github.com/primait/nuvola/pkg/connector/services/aws/lambda"
	"github.com/primait/nuvola/pkg/connector/services/aws/s3"
	neo4j "github.com/primait/nuvola/pkg/connector/services/neo4j"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/spf13/viper"
)

func NewStorageConnector() *StorageConnector {
	neo4jURL := viper.GetString("NEO4J_URL")
	neo4jPassword := viper.GetString("NEO4J_PASS")
	logger := logging.GetLogManager()
	client, err := neo4j.Connect(neo4jURL, "neo4j", neo4jPassword)
	if err != nil {
		logger.Error("Error connecting to database", "err", err)
	}
	connector := &StorageConnector{
		Client: *client,
		logger: logger,
	}
	return connector
}

func (sc *StorageConnector) FlushAll() *StorageConnector {
	sc.logger.Info("Flushing the database")
	sc.Client.DeleteAll()
	return sc
}

func (sc *StorageConnector) ImportResults(what string, content []byte) {
	var whoami = regexp.MustCompile(`^Whoami`)
	var credentialReport = regexp.MustCompile(`^CredentialReport`)
	var users = regexp.MustCompile(`^Users`)
	var groups = regexp.MustCompile(`^Groups`)
	var roles = regexp.MustCompile(`^Roles`)
	var buckets = regexp.MustCompile(`^Buckets`)
	var ec2s = regexp.MustCompile(`^EC2s`)
	var vpcs = regexp.MustCompile(`^VPCs`)
	var lambdas = regexp.MustCompile(`^Lambdas`)
	var rds = regexp.MustCompile(`^RDS`)
	var dynamodbs = regexp.MustCompile(`^DynamoDBs`)
	var redshiftdbs = regexp.MustCompile(`^RedshiftDBs`)

	sc.logger.Debug(fmt.Sprintf("Importing: %s", what))
	switch {
	case whoami.MatchString(what):
	case credentialReport.MatchString(what):
	case users.MatchString(what):
		contentStruct := []iam.User{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddUsers(&contentStruct)
	case groups.MatchString(what):
		contentStruct := []iam.Group{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddGroups(&contentStruct)
	case roles.MatchString(what):
		contentStruct := []iam.Role{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddRoles(&contentStruct)
		sc.Client.AddLinksToResourcesIAM()
	case buckets.MatchString(what):
		contentStruct := []s3.Bucket{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddBuckets(&contentStruct)
	case ec2s.MatchString(what):
		contentStruct := []ec2.Instance{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddEC2(&contentStruct)
	case vpcs.MatchString(what):
		contentStruct := ec2.VPC{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddVPC(&contentStruct)
	case lambdas.MatchString(what):
		contentStruct := []lambda.Lambda{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddLambda(&contentStruct)
	case rds.MatchString(what):
		contentStruct := database.RDS{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddRDS(&contentStruct)
	case dynamodbs.MatchString(what):
		contentStruct := []database.DynamoDB{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddDynamoDB(&contentStruct)
	case redshiftdbs.MatchString(what):
		contentStruct := []database.RedshiftDB{}
		_ = json.Unmarshal(content, &contentStruct)
		sc.Client.AddRedshift(&contentStruct)
	default:
		sc.logger.Error("Error importing data", "data", what)
	}
	sc.logger.Info(fmt.Sprintf("Imported: %s", what))
}

func (sc *StorageConnector) ImportBulkResults(content map[string]interface{}) {
	for k, v := range content {
		value, err := json.Marshal(v)
		if err != nil {
			sc.logger.Error("Error on marshalling data", "err", err)
		}
		sc.ImportResults(k, value)
	}
}

func (sc *StorageConnector) Query(query string, arguments map[string]interface{}) []map[string]interface{} {
	return sc.Client.Query(query, arguments)
}
