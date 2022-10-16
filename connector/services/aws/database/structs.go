package database

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
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
	Config aws.Config
}

type RDS struct {
	Clusters  []rdsTypes.DBCluster
	Instances []rdsTypes.DBInstance
}

type RDSClient struct {
	client *rds.Client
	Config aws.Config
}

type RedshiftDB struct {
	redshiftTypes.Cluster
}

type RedshiftClient struct {
	client *redshift.Client
	Config aws.Config
}

var re *awshttp.ResponseError
