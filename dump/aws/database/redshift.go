package database

import (
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	"github.com/aws/aws-sdk-go-v2/service/redshift/types"
)

// aws iam list-users
func ListRedshiftDBs(cfg awsconfig.AWSConfig) (redshiftDBs []*RedshiftDB) {
	var redshiftClient = RedshiftClient{Config: cfg}

	for i := range awsconfig.Regions {
		cfg.Config.Region = awsconfig.Regions[i]
		redshiftClient.client = redshift.NewFromConfig(cfg.Config)
		clusters := redshiftClient.listRedshiftClustersForRegion()
		for _, c := range clusters {
			var redDB *RedshiftDB = &RedshiftDB{
				Cluster: c,
			}
			redshiftDBs = append(redshiftDBs, redDB)
		}
	}

	return
}

func (rc *RedshiftClient) listRedshiftClustersForRegion() (clusters []types.Cluster) {
	var re *awshttp.ResponseError

	output, err := rc.client.DescribeClusters(context.TODO(), &redshift.DescribeClustersInput{})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			rc.Config.WaitAPILimit()
			return rc.listRedshiftClustersForRegion()
		} else if re.HTTPStatusCode() == 403 {
			return
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "Redshift", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	clusters = output.Clusters
	return
}
