package ec2

import (
	"context"
	b64 "encoding/base64"
	"errors"
	"log"
	"sync"

	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"golang.org/x/sync/semaphore"
)

func ListInstances(cfg awsconfig.AWSConfig) (EC2s []*Instance) {
	ec2Client = EC2Client{Config: cfg, client: ec2.NewFromConfig(cfg.Config)}

	for _, region := range awsconfig.Regions {
		cfg.Config.Region = region
		ec2Client.client = ec2.NewFromConfig(cfg.Config)
		EC2s = append(EC2s, ec2Client.listInstancesForRegion()...)
	}
	return
}

func (ec *EC2Client) listInstancesForRegion() (EC2s []*Instance) {
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(20)) // TODO: parametric
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
		if re.HTTPStatusCode() == 503 {
			ec.Config.WaitAPILimit()
			return ec.listInstancesForRegion()
		} else {
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	for _, instances := range output.Reservations {
		wg.Add(1)
		go func(instances types.Reservation) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v\n", err)
			}
			defer sem.Release(1)
			defer wg.Done()
			var instancesSlice []*Instance
			for _, instance := range instances.Instances {
				instancesSlice = append(instancesSlice, &Instance{
					Instance:          instance,
					UserData:          ec.getInstanceUserDataAttribute(aws.ToString(instance.InstanceId)),
					NetworkInterfaces: ec.getNetworkInterfacesWithGroups(instance.NetworkInterfaces),
					InstanceState:     ec.getInstanceState(aws.ToString(instance.InstanceId)),
				})
			}
			mu.Lock()
			defer mu.Unlock()
			EC2s = append(EC2s, instancesSlice...)
		}(instances)
	}
	wg.Wait()
	return
}

func (ec *EC2Client) getInstanceUserDataAttribute(instanceId string) string {
	var (
		decodedData []byte
	)

	userData, err := ec.client.DescribeInstanceAttribute(context.TODO(), &ec2.DescribeInstanceAttributeInput{
		InstanceId: &instanceId,
		Attribute:  types.InstanceAttributeNameUserData,
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			ec.Config.WaitAPILimit()
			return ec.getInstanceUserDataAttribute(instanceId)
		} else {
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
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

func (ec *EC2Client) getSecurityGroups(groupId string) (secGroups []types.SecurityGroup) {
	output, err := ec.client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
		GroupIds: []string{groupId},
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			ec.Config.WaitAPILimit()
			return ec.getSecurityGroups(groupId)
		} else {
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	secGroups = append(secGroups, output.SecurityGroups...)
	return
}

func (ec *EC2Client) getInstanceState(instanceId string) (state types.InstanceState) {
	output, err := ec.client.DescribeInstanceStatus(context.TODO(), &ec2.DescribeInstanceStatusInput{
		InstanceIds: []string{instanceId},
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 503 {
			ec.Config.WaitAPILimit()
			return ec.getInstanceState(instanceId)
		} else {
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	if output != nil && len(output.InstanceStatuses) > 0 {
		return *output.InstanceStatuses[0].InstanceState
	} else {
		return
	}
}
