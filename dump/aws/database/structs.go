package database

import (
	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsTypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	redshiftTypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
)

type DynamoDB struct {
	Table
}

type Table struct {
	Name   string
	Region string
}

type DynamoClient struct {
	client *dynamodb.Client
	Config awsconfig.AWSConfig
}

type RDS struct {
	Clusters  []rdsTypes.DBCluster
	Instances []rdsTypes.DBInstance
}

type RDSClient struct {
	client *rds.Client
	Config awsconfig.AWSConfig
}

type RedshiftDB struct {
	redshiftTypes.Cluster
}

type RedshiftClient struct {
	client *redshift.Client
	Config awsconfig.AWSConfig
}
