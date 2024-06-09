package iam

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"reflect"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/sourcegraph/conc/iter"
)

// aws iam list-roles and aws iam list-instance-profiles
func ListRoles(cfg aws.Config) (roles []*Role) {
	iamClient = IAMClient{Config: cfg, client: iam.NewFromConfig(cfg), logger: logging.GetLogManager()}

	roles = iter.Map(iamClient.listRoles(), func(role *types.Role) *Role {
		var assumeRoleDocument = PolicyDocument{}
		var instanceProfileRef = ""
		var instanceProfileArn = ""
		decodedValue, _ := url.QueryUnescape(*role.AssumeRolePolicyDocument)
		err := json.Unmarshal([]byte(decodedValue), &assumeRoleDocument)
		if err != nil {
			iamClient.logger.Warn("Error unmarshalling assumeRoleDocument", "err", err)
		}

		// Sort Service object in the AssumeRolePolicyDocument; useful to diff different JSON outputs
		assumableBy := []string{}
		for i, assStatement := range assumeRoleDocument.Statement {
			if reflect.ValueOf(assStatement.Principal.Service).Kind() == reflect.String {
				assumableBy = append(assumableBy, assumeRoleDocument.Statement[i].Principal.Service.(string))
			} else if reflect.ValueOf(assStatement.Principal.Service).Kind() == reflect.Slice {
				assumeRoleDocument.Statement[i].Principal.Service = sortStringSlice(assStatement.Principal.Service)
				assumableBy = append(assumableBy, assumeRoleDocument.Statement[i].Principal.Service.([]string)...)
			}

			if reflect.ValueOf(assStatement.Principal.AWS).Kind() == reflect.String {
				assumableBy = append(assumableBy, assumeRoleDocument.Statement[i].Principal.AWS.(string))
			} else if reflect.ValueOf(assStatement.Principal.AWS).Kind() == reflect.Slice {
				assumeRoleDocument.Statement[i].Principal.AWS = sortStringSlice(assStatement.Principal.AWS)
				assumableBy = append(assumableBy, assumeRoleDocument.Statement[i].Principal.AWS.([]string)...)
			}

			if reflect.ValueOf(assStatement.Principal.Federated).Kind() == reflect.String {
				assumableBy = append(assumableBy, assumeRoleDocument.Statement[i].Principal.Federated.(string))
			} else if reflect.ValueOf(assStatement.Principal.Federated).Kind() == reflect.Slice {
				assumeRoleDocument.Statement[i].Principal.Federated = sortStringSlice(assStatement.Principal.Federated)
				assumableBy = append(assumableBy, assumeRoleDocument.Statement[i].Principal.Federated.([]string)...)
			}
		}
		sort.Strings(assumableBy)

		for _, instanceProfile := range iamClient.listInstanceProfiles() {
			for _, r := range instanceProfile.Roles {
				if aws.ToString(r.RoleId) == aws.ToString(role.RoleId) {
					instanceProfileRef = aws.ToString(instanceProfile.InstanceProfileId)
					instanceProfileArn = aws.ToString(instanceProfile.Arn)
				}
			}
		}

		inline := iamClient.listInlinePolicies(aws.ToString(role.RoleName), "role")
		attached := iamClient.listAttachedPolicies(aws.ToString(role.RoleName), "role")
		return &Role{
			Role:                     *role,
			AssumeRolePolicyDocument: assumeRoleDocument,
			AssumableBy:              assumableBy,
			AttachedPolicies:         attached,
			InlinePolicies:           inline,
			InstanceProfileID:        instanceProfileRef,
			InstanceProfileArn:       instanceProfileArn,
		}
	})

	sort.Slice(roles, func(i, j int) bool {
		return aws.ToString(roles[i].RoleName) < aws.ToString(roles[j].RoleName)
	})

	return
}

func (ic *IAMClient) listRoles() []types.Role {
	var (
		marker         *string
		collectedRoles []types.Role
	)

	for {
		roleOutput, err := iamClient.client.ListRoles(context.TODO(), &iam.ListRolesInput{
			Marker:   marker,
			MaxItems: aws.Int32(300),
		})
		if errors.As(err, &re) {
			ic.logger.Warn("Error on ListRoles", "err", re)
		}

		collectedRoles = append(collectedRoles, roleOutput.Roles...)
		if !roleOutput.IsTruncated {
			break
		}
		marker = roleOutput.Marker
	}
	return collectedRoles
}

func (ic *IAMClient) listInstanceProfiles() []types.InstanceProfile {
	var (
		marker                    *string
		collectedInstanceProfiles []types.InstanceProfile
	)

	for {
		roleOutput, err := iamClient.client.ListInstanceProfiles(context.TODO(), &iam.ListInstanceProfilesInput{
			Marker:   marker,
			MaxItems: aws.Int32(300),
		})
		if errors.As(err, &re) {
			ic.logger.Warn("Error on ListInstanceProfiles", "err", re)
		}

		collectedInstanceProfiles = append(collectedInstanceProfiles, roleOutput.InstanceProfiles...)
		if !roleOutput.IsTruncated {
			break
		}
		marker = roleOutput.Marker
	}
	return collectedInstanceProfiles
}
