package iam

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/gocarina/gocsv"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/sourcegraph/conc/iter"
)

// aws iam list-users
func ListUsers(cfg aws.Config, credentialReport map[string]*CredentialReport) (users []*User) {
	if len(credentialReport) > 0 {
		rootAccount := credentialReport["<root_account>"]
		rootDate, _ := time.Parse("2006-01-02T15:04:05+00:00", rootAccount.UserCreation)
		rootUsedDate, _ := time.Parse("2006-01-02T15:04:05+00:00", rootAccount.PasswordLastUsed)
		users = append(users, &User{
			User: types.User{
				UserName:         &rootAccount.User,
				Arn:              &rootAccount.Arn,
				CreateDate:       &rootDate,
				PasswordLastUsed: &rootUsedDate,
				UserId:           aws.String("0"),
			},
			PasswordEnabled:     rootAccount.PasswordEnabled,
			PasswordLastChanged: rootAccount.PasswordLastChanged,
			MfaActive:           rootAccount.MfaActive,
		})
	}

	users = append(users, iter.Map(listUsers(), func(user *types.User) *User {
		groups := iamClient.listGroupsForUser(aws.ToString(user.UserName))
		inline := iamClient.listInlinePolicies(aws.ToString(user.UserName), "user")
		attached := iamClient.listAttachedPolicies(aws.ToString(user.UserName), "user")
		accessKeys := iamClient.listAccessKeys(aws.ToString(user.UserName))
		loginProfile := iamClient.listLoginProfile(aws.ToString(user.UserName))

		userAccount := credentialReport[aws.ToString(user.UserName)]
		return &User{
			User:                *user,
			Groups:              groups,
			AccessKeys:          accessKeys,
			LoginProfile:        loginProfile,
			InlinePolicies:      inline,
			AttachedPolicies:    attached,
			PasswordEnabled:     userAccount.PasswordEnabled,
			PasswordLastChanged: userAccount.PasswordLastChanged,
			MfaActive:           userAccount.MfaActive,
		}
	})...)

	sort.Slice(users, func(i, j int) bool {
		return aws.ToString(users[i].UserName) < aws.ToString(users[j].UserName)
	})

	return
}

func listUsers() (collectedUsers []types.User) {
	var (
		marker *string
	)

	for {
		output, err := iamClient.client.ListUsers(context.TODO(), &iam.ListUsersInput{
			Marker: marker,
		})
		if errors.As(err, &re) {
			logging.HandleAWSError(re, "IAM - Users", "ListUsers")
		}

		collectedUsers = append(collectedUsers, output.Users...)
		if !output.IsTruncated {
			break
		}
		marker = output.Marker
	}
	return collectedUsers
}

// aws iam get-credential-report
func GetCredentialReport(cfg aws.Config) (credentialReport map[string]*CredentialReport) {
	var (
		countRetries = 0
		maxRetries   = 5
	)

	iamClient := iam.NewFromConfig(cfg)
	output, err := iamClient.GetCredentialReport(context.TODO(), &iam.GetCredentialReportInput{})

	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 410 { // Gone: https://http.cat/410
			checkGen, err := iamClient.GenerateCredentialReport(context.TODO(), &iam.GenerateCredentialReportInput{})
			if errors.As(err, &re) {
				logging.HandleAWSError(re, "IAM - Users", "GenerateCredentialReport")
			}
			log.Println("Credential Report generation requested...")
			for checkGen.State != "COMPLETE" {
				if countRetries >= maxRetries {
					logging.HandleAWSError(re, "IAM - Policies", "GenerateCredentialReport")
				}
				countRetries++
				time.Sleep(5 * time.Second)
				checkGen, err = iamClient.GenerateCredentialReport(context.TODO(), &iam.GenerateCredentialReportInput{})
				if errors.As(err, &re) {
					logging.HandleAWSError(re, "IAM - Users", "GenerateCredentialReport")
				}
			}
			return GetCredentialReport(cfg)
		} else {
			logging.HandleAWSError(re, "IAM - Users", "GetCredentialReport")
		}
		return nil
	}

	credentialReportCSV := []*CredentialReport{}
	if err := gocsv.Unmarshal(bytes.NewReader(output.Content), &credentialReportCSV); err != nil {
		logging.HandleError(err, "IAM - Users", "Umarshalling credentialReportCSV")
	}

	credentialReport = make(map[string]*CredentialReport)
	for i := range credentialReportCSV {
		credentialReport[credentialReportCSV[i].User] = credentialReportCSV[i]
	}
	return
}

func (ic *IAMClient) listAccessKeys(identity string) (accessKeys []types.AccessKeyMetadata) {
	output, err := ic.client.ListAccessKeys(context.TODO(), &iam.ListAccessKeysInput{
		UserName: &identity,
	})
	if errors.As(err, &re) {
		logging.HandleAWSError(re, "IAM - Users", "ListAccessKeys")
	}
	accessKeys = output.AccessKeyMetadata
	return
}

func (ic *IAMClient) listLoginProfile(identity string) (loginProfile types.LoginProfile) {
	output, err := ic.client.GetLoginProfile(context.TODO(), &iam.GetLoginProfileInput{
		UserName: &identity,
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() != 404 { // an user may not have a login profile
			logging.HandleAWSError(re, "IAM - Users", "GetLoginProfile")
		}
		return
	}

	loginProfile = *output.LoginProfile
	return
}