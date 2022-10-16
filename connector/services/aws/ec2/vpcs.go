package ec2

import (
	"context"
	"errors"
	nuvolaerror "nuvola/tools/error"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func ListVpcs(cfg aws.Config) (vpcs *VPC) {
	ec2Client := EC2Client{Config: cfg, client: ec2.NewFromConfig(cfg)}

	vpcs = &VPC{}
	for _, region := range Regions {
		cfg.Region = region
		ec2Client.client = ec2.NewFromConfig(cfg)
		vpcsAndPeerings := ec2Client.getVpcs()
		vpcs.Peerings = append(vpcs.Peerings, vpcsAndPeerings.Peerings...)
		vpcs.VPCs = append(vpcs.VPCs, vpcsAndPeerings.VPCs...)
	}
	return
}

func (ec *EC2Client) getVpcs() (vpcs *VPC) {
	vpcs = &VPC{}

	vpcsOutput, err := ec.client.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{
		MaxResults: aws.Int32(1000),
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "EC2 - VPC", "DescribeVpcs")
	}

	peeringOutput, err := ec.client.DescribeVpcPeeringConnections(context.TODO(), &ec2.DescribeVpcPeeringConnectionsInput{
		MaxResults: aws.Int32(1000),
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "EC2 - VPC", "DescribeVpcPeeringConnections")
	}

	for i := 0; i < len(vpcsOutput.Vpcs); i++ {
		vpcs.VPCs = append(vpcs.VPCs, vpcsOutput.Vpcs[i])
	}

	for i := 0; i < len(peeringOutput.VpcPeeringConnections); i++ {
		vpcs.Peerings = append(vpcs.Peerings, peeringOutput.VpcPeeringConnections[i])
	}

	return
}
