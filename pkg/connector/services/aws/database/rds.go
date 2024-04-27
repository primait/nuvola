package database

import (
	"context"
	"errors"

	"github.com/primait/nuvola/pkg/connector/services/aws/ec2"
	"github.com/primait/nuvola/pkg/io/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// aws iam list-users
func ListRDS(cfg aws.Config) (rdsRet *RDS, re *awshttp.ResponseError) {
	var rdsClient = RDSClient{Config: cfg}

	rdsRet = &RDS{}

	for i := range ec2.Regions {
		cfg.Region = ec2.Regions[i]
		rdsClient.client = rds.NewFromConfig(cfg)

		clusters := rdsClient.listRDSClustersForRegion()
		instances := rdsClient.listRDSInstancesForRegion()

		rdsRet.Clusters = append(rdsRet.Clusters, clusters...)
		rdsRet.Instances = append(rdsRet.Instances, instances...)
	}

	return
}

func (rc *RDSClient) listRDSClustersForRegion() (clusters []types.DBCluster) {
	output, err := rc.client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{})
	if errors.As(err, &re) {
		if re.Response.StatusCode != 501 { // When using LocalStack: this is a Pro feature
			logging.HandleAWSError(re, "RDS", "DescribeDBClusters")
		}
	}

	if output != nil {
		for i := 0; i < len(output.DBClusters); i++ {
			clusters = append(clusters, output.DBClusters[i])
		}
	}

	return
}

func (rc *RDSClient) listRDSInstancesForRegion() (instances []types.DBInstance) {
	output, err := rc.client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{})
	if errors.As(err, &re) {
		if re.Response.StatusCode != 501 { // When using LocalStack: this is a Pro feature
			logging.HandleAWSError(re, "RDS", "DescribeDBInstances")
		}
	}

	if output != nil {
		for i := 0; i < len(output.DBInstances); i++ {
			instances = append(instances, output.DBInstances[i])
		}
	}
	return
}
