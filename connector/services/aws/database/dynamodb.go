package database

import (
	"context"
	"errors"

	"github.com/primait/nuvola/connector/services/aws/ec2"
	"github.com/primait/nuvola/pkg/io/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// aws iam list-users
func ListDynamoDBs(cfg aws.Config) (dynamoDBs []*DynamoDB) {
	var dynamoClient = DynamoClient{Config: cfg}

	for i := range ec2.Regions {
		cfg.Region = ec2.Regions[i]
		dynamoClient.client = dynamodb.NewFromConfig(cfg)

		tables := dynamoClient.listDynamoDBTablesForRegion()
		for _, table := range tables {
			var db = &DynamoDB{Table{
				Name:   table,
				Region: ec2.Regions[i],
			}}
			dynamoDBs = append(dynamoDBs, db)
		}
	}

	return
}

func (dc *DynamoClient) listDynamoDBTablesForRegion() (tableNames []string) {
	output, err := dc.client.ListTables(context.TODO(), &dynamodb.ListTablesInput{
		Limit: aws.Int32(100),
	})
	if errors.As(err, &re) {
		logging.HandleAWSError(re, "DynamoDB", "ListTables")
	}

	tableNames = output.TableNames
	return
}
