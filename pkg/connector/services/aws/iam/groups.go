package iam

import (
	"context"
	"errors"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/sourcegraph/conc/iter"
)

func ListGroups(cfg aws.Config) (groups []*Group) {
	iamClient = IAMClient{client: iam.NewFromConfig(cfg), Config: cfg, logger: logging.GetLogManager()}

	groups = iter.Map(iamClient.listGroups(), func(group *types.Group) *Group {
		inlines := iamClient.listInlinePolicies(aws.ToString(group.GroupName), "group")
		attached := iamClient.listAttachedPolicies(aws.ToString(group.GroupName), "group")

		return &Group{
			Group:            *group,
			InlinePolicies:   inlines,
			AttachedPolicies: attached,
		}
	})

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
		ic.logger.Warn("Error on ListGroupsForUser", "err", re)
	}
	return output.Groups
}

func (ic *IAMClient) listGroups() (collectedGroups []types.Group) {
	var marker *string
	for {
		output, err := iamClient.client.ListGroups(context.TODO(), &iam.ListGroupsInput{
			Marker: marker,
		})
		if errors.As(err, &re) {
			ic.logger.Warn("Error on ListGroups", "err", re)
		}

		collectedGroups = append(collectedGroups, output.Groups...)
		if !output.IsTruncated {
			break
		}
		marker = output.Marker
	}
	return
}
