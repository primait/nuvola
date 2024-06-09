package ec2

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/primait/nuvola/pkg/io/logging"
)

type EC2Client struct {
	client *ec2.Client
	aws.Config
	logger logging.LogManager
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
	re      *awshttp.ResponseError
	Regions []string
)

func ListAndSaveRegions(cfg aws.Config) {
	if len(Regions) == 0 {
		ec2Client := ec2.NewFromConfig(cfg)

		output, err := ec2Client.DescribeRegions(context.TODO(), &ec2.DescribeRegionsInput{})
		if errors.As(err, &re) || output == nil {
			logging.GetLogManager().Warn("Error on listing regions", "err", err)
		} else {
			for _, region := range output.Regions {
				Regions = append(Regions, aws.ToString(region.RegionName))
			}
		}
	}
}
