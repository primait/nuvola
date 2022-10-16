package ec2

import (
	"context"
	b64 "encoding/base64"
	"errors"
	nuvolaerror "nuvola/tools/error"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"golang.org/x/sync/semaphore"
)

func ListInstances(cfg aws.Config) (ec2s []*Instance, err *awshttp.ResponseError) {
	ec2Client := EC2Client{Config: cfg, client: ec2.NewFromConfig(cfg)}

	for _, region := range Regions {
		cfg.Region = region
		ec2Client.client = ec2.NewFromConfig(cfg)
		instances := ec2Client.listInstancesForRegion()
		ec2s = append(ec2s, instances...)
	}
	return
}

func (ec *EC2Client) listInstancesForRegion() (ec2s []*Instance) {
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(15))
		wg  sync.WaitGroup
	)

	output, err := ec.client.DescribeInstances(context.TODO(), &ec2.DescribeInstancesInput{
		MaxResults: aws.Int32(1000),
		Filters: []types.Filter{{
			Name:   aws.String("instance-state-name"),
			Values: []string{"running", "pending"},
		}},
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "EC2", "listInstancesForRegion")
	}

	for _, instances := range output.Reservations {
		wg.Add(1)
		go func(instances types.Reservation) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				nuvolaerror.HandleError(err, "EC2", "listInstancesForRegion - Acquire Semaphore")
			}
			defer sem.Release(1)
			defer wg.Done()
			var instancesSlice []*Instance
			for _, instance := range instances.Instances {
				userData := ec.getInstanceUserDataAttribute(aws.ToString(instance.InstanceId))
				instancesSlice = append(instancesSlice, &Instance{
					Instance:          instance,
					UserData:          userData,
					NetworkInterfaces: ec.getNetworkInterfacesWithGroups(instance.NetworkInterfaces),
					InstanceState:     ec.getInstanceState(aws.ToString(instance.InstanceId)),
				})
			}
			mu.Lock()
			defer mu.Unlock()
			ec2s = append(ec2s, instancesSlice...)
		}(instances)
	}
	wg.Wait()
	return
}

func (ec *EC2Client) getInstanceUserDataAttribute(instanceID string) string {
	var decodedData []byte

	userData, err := ec.client.DescribeInstanceAttribute(context.TODO(), &ec2.DescribeInstanceAttributeInput{
		InstanceId: &instanceID,
		Attribute:  types.InstanceAttributeNameUserData,
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "EC2", "DescribeInstanceAttribute")
	}

	if userData.UserData != nil {
		decodedData, _ = b64.StdEncoding.DecodeString(aws.ToString(userData.UserData.Value))
	}
	return string(decodedData)
}

func (ec *EC2Client) getNetworkInterfacesWithGroups(netInts []types.InstanceNetworkInterface) (output []NetworkInterface) {
	for _, netInt := range netInts {
		itemNetInt := NetworkInterface{
			InstanceNetworkInterface: netInt,
		}
		for _, group := range netInt.Groups {
			itemNetInt.SecurityGroup = append(itemNetInt.SecurityGroup, ec.getSecurityGroups(*group.GroupId)...)
		}
		output = append(output, itemNetInt)
	}
	return
}

func (ec *EC2Client) getSecurityGroups(groupID string) (secGroups []types.SecurityGroup) {
	output, err := ec.client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupID},
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "EC2", "DescribeSecurityGroups")
	}

	secGroups = append(secGroups, output.SecurityGroups...)
	return
}

func (ec *EC2Client) getInstanceState(instanceID string) (state types.InstanceState) {
	output, err := ec.client.DescribeInstanceStatus(context.TODO(), &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{instanceID},
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "EC2", "DescribeSecurityGroups")
	}

	if output != nil && len(output.InstanceStatuses) > 0 {
		return *output.InstanceStatuses[0].InstanceState
	}
	return
}
