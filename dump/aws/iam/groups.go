package iam

import (
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"golang.org/x/sync/semaphore"
)

func (ic *IAMClient) listGroupsForUser(identity string) (groups []types.Group) {
	output, err := ic.client.ListGroupsForUser(context.TODO(), &iam.ListGroupsForUserInput{
		UserName: &identity,
	})
	if errors.As(err, &re) {
		log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - groups", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}
	groups = output.Groups
	return
}

func ListGroups(cfg awsconfig.AWSConfig) (groups []*Group) {
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(20)) // TODO: parametric
		wg  sync.WaitGroup
	)

	for _, group := range listGroups() {
		wg.Add(1)
		go func(group types.Group) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v\n", err)
			}
			defer sem.Release(1)
			defer mu.Unlock()
			defer wg.Done()
			var item *Group = &Group{
				Group:            group,
				InlinePolicies:   iamClient.listInlinePolicies(aws.ToString(group.GroupName), "group"),
				AttachedPolicies: iamClient.listAttachedPolicies(aws.ToString(group.GroupName), "group"),
			}
			mu.Lock()
			groups = append(groups, item)
		}(group)
	}
	wg.Wait()

	sort.Slice(groups, func(i, j int) bool {
		return aws.ToString(groups[i].GroupName) < aws.ToString(groups[j].GroupName)
	})

	return
}

func listGroups() []types.Group {
	var (
		marker          *string
		collectedGroups []types.Group
	)

	for {
		output, err := iamClient.client.ListGroups(context.TODO(), &iam.ListGroupsInput{
			Marker: marker,
		})
		if errors.As(err, &re) {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - users", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}

		collectedGroups = append(collectedGroups, output.Groups...)
		if !output.IsTruncated {
			break
		}
		marker = output.Marker
	}
	return collectedGroups
}
