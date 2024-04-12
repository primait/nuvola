package database

import (
	"context"
	"errors"

	nuvolaerror "github.com/primait/nuvola/tools/error"

	"github.com/primait/nuvola/connector/services/aws/ec2"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/aws/aws-sdk-go-v2/service/redshift/types"
)

// aws iam list-users
func ListRedshiftDBs(cfg aws.Config) (redshiftDBs []*RedshiftDB) {
	var redshiftClient = RedshiftClient{Config: cfg}

	for i := range ec2.Regions {
		cfg.Region = ec2.Regions[i]
		redshiftClient.client = redshift.NewFromConfig(cfg)
		clusters := redshiftClient.listRedshiftClustersForRegion()
		for _, c := range clusters {
			var redDB = &RedshiftDB{
				Cluster: c,
			}
			redshiftDBs = append(redshiftDBs, redDB)
		}
	}

	return
}

func (rc *RedshiftClient) listRedshiftClustersForRegion() (clusters []types.Cluster) {
	output, err := rc.client.DescribeClusters(context.TODO(), &redshift.DescribeClustersInput{})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "Redshift", "DescribeClusters")
	}

	clusters = output.Clusters
	return
}
