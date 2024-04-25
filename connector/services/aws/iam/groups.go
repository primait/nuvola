package iam

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/primait/nuvola/pkg/io/logging"
	"golang.org/x/sync/semaphore"
)

func ListGroups(cfg aws.Config) (groups []*Group) {
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(30))
		wg  sync.WaitGroup
	)

	iamClient = IAMClient{client: iam.NewFromConfig(cfg), Config: cfg}
	for _, group := range listGroups() {
		wg.Add(1)
		go func(group types.Group) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				logging.HandleError(err, "IAM - Groups", "ListGroups - Acquire Semaphore")
			}
			defer sem.Release(1)
			defer mu.Unlock()
			defer wg.Done()

			inlines := iamClient.listInlinePolicies(aws.ToString(group.GroupName), "group")
			attached := iamClient.listAttachedPolicies(aws.ToString(group.GroupName), "group")

			var item = &Group{
				Group:            group,
				InlinePolicies:   inlines,
				AttachedPolicies: attached,
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

func (ic *IAMClient) listGroupsForUser(identity string) []types.Group {
	output, err := ic.client.ListGroupsForUser(context.TODO(), &iam.ListGroupsForUserInput{
		UserName: &identity,
	})
	if errors.As(err, &re) {
		logging.HandleAWSError(re, "IAM - Groups", "ListGroupsForUser")
	}
	return output.Groups
}

func listGroups() (collectedGroups []types.Group) {
	var marker *string
	for {
		output, err := iamClient.client.ListGroups(context.TODO(), &iam.ListGroupsInput{
			Marker: marker,
		})
		if errors.As(err, &re) {
			logging.HandleAWSError(re, "IAM - Groups", "ListGroups")
		}

		collectedGroups = append(collectedGroups, output.Groups...)
		if !output.IsTruncated {
			break
		}
		marker = output.Marker
	}
	return
}
