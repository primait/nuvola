package iam

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	awsconfig "nuvola/config/aws"
	"reflect"
	"sort"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"golang.org/x/sync/semaphore"
)

// aws iam list-roles and aws iam list-instance-profiles
func ListRoles(cfg awsconfig.AWSConfig) (roles []*Role) {
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(18)) // TODO: parametric
		wg  sync.WaitGroup
	)
	iamClient = IAMClient{Config: cfg, client: iam.NewFromConfig(cfg.Config)}

	for _, role := range listRoles() {
		wg.Add(1)
		go func(role types.Role) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v\n", err)
			}

			defer sem.Release(1)
			defer mu.Unlock()
			defer wg.Done()
			var assumeRoleDocument = awsconfig.PolicyDocument{}
			var instanceProfileRef = ""
			var instanceProfileArn = ""
			decodedValue, _ := url.QueryUnescape(*role.AssumeRolePolicyDocument)
			err := json.Unmarshal([]byte(decodedValue), &assumeRoleDocument)
			if err != nil {
				log.Fatalln(err)
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

			for _, instanceProfile := range listInstanceProfiles() {
				for _, r := range instanceProfile.Roles {
					if aws.ToString(r.RoleId) == aws.ToString(role.RoleId) {
						instanceProfileRef = aws.ToString(instanceProfile.InstanceProfileId)
						instanceProfileArn = aws.ToString(instanceProfile.Arn)
					}
				}
			}

			var item *Role = &Role{
				Role:                     role,
				AssumeRolePolicyDocument: assumeRoleDocument,
				AssumableBy:              assumableBy,
				AttachedPolicies:         iamClient.listAttachedPolicies(aws.ToString(role.RoleName), "role"),
				InlinePolicies:           iamClient.listInlinePolicies(aws.ToString(role.RoleName), "role"),
				InstanceProfileId:        instanceProfileRef,
				InstanceProfileArn:       instanceProfileArn,
			}
			mu.Lock()
			roles = append(roles, item)
		}(role)
	}
	wg.Wait()

	sort.Slice(roles, func(i, j int) bool {
		return aws.ToString(roles[i].RoleName) < aws.ToString(roles[j].RoleName)
	})

	return
}

func listRoles() []types.Role {
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
			if re.HTTPStatusCode() == 429 || re.HTTPStatusCode() == 400 { // Too Many Requests https://http.cat/429; Throttling: Rate exceeded
				iamClient.Config.WaitAPILimit()
				continue
			} else {
				log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - roles", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
			}
		}

		collectedRoles = append(collectedRoles, roleOutput.Roles...)
		if !roleOutput.IsTruncated {
			break
		}
		marker = roleOutput.Marker
	}
	return collectedRoles
}

func listInstanceProfiles() []types.InstanceProfile {
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
			if re.HTTPStatusCode() == 429 || re.HTTPStatusCode() == 400 { // Too Many Requests https://http.cat/429; Throttling: Rate exceeded
				iamClient.Config.WaitAPILimit()
				continue
			} else {
				log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - instance profiles", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
			}
		}

		collectedInstanceProfiles = append(collectedInstanceProfiles, roleOutput.InstanceProfiles...)
		if !roleOutput.IsTruncated {
			break
		}
		marker = roleOutput.Marker
	}
	return collectedInstanceProfiles
}
