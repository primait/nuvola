package ec2

import (
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func ListVpcs(cfg awsconfig.AWSConfig) (VPCs *VPC) {
	ec2Client = EC2Client{Config: cfg, client: ec2.NewFromConfig(cfg.Config)}

	VPCs = &VPC{}
	for _, region := range awsconfig.Regions {
		cfg.Config.Region = region
		ec2Client.client = ec2.NewFromConfig(cfg.Config)
		vpcs := ec2Client.getVpcs()
		VPCs.Peerings = append(VPCs.Peerings, vpcs.Peerings...)
		VPCs.VPCs = append(VPCs.VPCs, vpcs.VPCs...)
	}
	return
}

func (ec *EC2Client) getVpcs() (vpcs *VPC) {
	vpcs = &VPC{}

	vpcsOutput, err := ec.client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
		MaxResults: aws.Int32(1000),
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			ec.Config.WaitAPILimit()
			return ec.getVpcs()
		} else {
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	peeringOutput, err := ec.client.DescribeVpcPeeringConnections(context.TODO(), &ec2.DescribeVpcPeeringConnectionsInput{
		MaxResults: aws.Int32(1000),
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			ec.Config.WaitAPILimit()
			return ec.getVpcs()
		} else {
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	for i := 0; i < len(vpcsOutput.Vpcs); i++ {
		vpcs.VPCs = append(vpcs.VPCs, vpcsOutput.Vpcs[i])
	}

	for i := 0; i < len(peeringOutput.VpcPeeringConnections); i++ {
		vpcs.Peerings = append(vpcs.Peerings, peeringOutput.VpcPeeringConnections[i])
	}

	return
}
