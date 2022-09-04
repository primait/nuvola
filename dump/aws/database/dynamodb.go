package database

import (
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// aws iam list-users
func ListDynamoDBs(cfg awsconfig.AWSConfig) (dynamoDBs []*DynamoDB) {
	var dynamoClient = DynamoClient{Config: cfg}

	for i := range awsconfig.Regions {
		cfg.Config.Region = awsconfig.Regions[i]
		dynamoClient.client = dynamodb.NewFromConfig(cfg.Config)

		for _, table := range dynamoClient.listDynamoDBTablesForRegion() {
			var db *DynamoDB = &DynamoDB{Table{
				Name:   table,
				Region: awsconfig.Regions[i],
			}}
			dynamoDBs = append(dynamoDBs, db)
		}
	}

	return
}

func (dc *DynamoClient) listDynamoDBTablesForRegion() (tableNames []string) {
	var re *awshttp.ResponseError

	output, err := dc.client.ListTables(context.TODO(), &dynamodb.ListTablesInput{})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			dc.Config.WaitAPILimit()
			return dc.listDynamoDBTablesForRegion()
		} else if re.HTTPStatusCode() == 403 {
			return
		} else if re.HTTPStatusCode() == 400 {
			log.Printf("Service: %s, Region: %s, error: %v\n", "DynamoDB", dc.Config.Config.Region, re.Unwrap())
			return
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "DynamoDB", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	tableNames = output.TableNames
	return
}
