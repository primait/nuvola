package ec2

import (
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2Client struct {
	client *ec2.Client
	Config awsconfig.AWSConfig
}

// Override SDK EC2 instance type to insert SecurityGroup information
type Instance struct {
	types.Instance
	UserData          string `json:"UserData,omitempty"`
	NetworkInterfaces []NetworkInterface
	InstanceState     types.InstanceState
}

type VPC struct {
	VPCs     []types.Vpc
	Peerings []types.VpcPeeringConnection
}

type NetworkInterface struct {
	types.InstanceNetworkInterface
	SecurityGroup []types.SecurityGroup
}

var (
	re        *awshttp.ResponseError
	ec2Client EC2Client
)

func ListAndSaveRegions(cfg awsconfig.AWSConfig) {
	if len(awsconfig.Regions) <= 0 {
		ec2Client = EC2Client{Config: cfg, client: ec2.NewFromConfig(cfg.Config)}

		output, err := ec2Client.client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
		if errors.As(err, &re) {
			if re.HTTPStatusCode() == 503 {
				cfg.WaitAPILimit()
				ListAndSaveRegions(cfg)
			} else {
				log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
			}
		}

		for _, region := range output.Regions {
			awsconfig.Regions = append(awsconfig.Regions, aws.ToString(region.RegionName))
		}
	}
}
