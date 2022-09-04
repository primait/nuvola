package database

import (
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// aws iam list-users
func ListRDS(cfg awsconfig.AWSConfig) (rdsRet *RDS) {
	var rdsClient = RDSClient{Config: cfg}

	rdsRet = &RDS{}

	for i := range awsconfig.Regions {
		cfg.Config.Region = awsconfig.Regions[i]
		rdsClient.client = rds.NewFromConfig(cfg.Config)
		rdsRet.Clusters = append(rdsRet.Clusters, rdsClient.listRDSClustersForRegion()...)
		rdsRet.Instances = append(rdsRet.Instances, rdsClient.listRDSInstancesForRegion()...)
	}

	return
}

func (rc *RDSClient) listRDSClustersForRegion() (clusters []types.DBCluster) {
	var re *awshttp.ResponseError

	output, err := rc.client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			rc.Config.WaitAPILimit()
			return rc.listRDSClustersForRegion()
		} else if re.HTTPStatusCode() == 403 {
			return
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "RDS Clusters", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	for i := 0; i < len(output.DBClusters); i++ {
		clusters = append(clusters, output.DBClusters[i])
	}

	return
}

func (rc *RDSClient) listRDSInstancesForRegion() (instances []types.DBInstance) {
	var re *awshttp.ResponseError

	output, err := rc.client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			rc.Config.WaitAPILimit()
			return rc.listRDSInstancesForRegion()
		} else if re.HTTPStatusCode() == 403 {
			return
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "RDS Instances", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	for i := 0; i < len(output.DBInstances); i++ {
		instances = append(instances, output.DBInstances[i])
	}
	return
}
