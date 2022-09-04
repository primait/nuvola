package iam

import (
	"bytes"
	"context"
	"errors"
	"log"
	awsconfig "nuvola/config/aws"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/gocarina/gocsv"
	"golang.org/x/sync/semaphore"
)

// aws iam list-users
func ListUsers(cfg awsconfig.AWSConfig, credentialReport map[string]*CredentialReport) (users []*User) {
	var (
		mu  = &sync.Mutex{}
		sem = semaphore.NewWeighted(int64(20)) // TODO: parametric
		wg  sync.WaitGroup
	)

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

	for _, user := range listUsers() {
		wg.Add(1)
		go func(user types.User) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v\n", err)
			}
			defer sem.Release(1)
			defer mu.Unlock()
			defer wg.Done()
			userAccount := credentialReport[aws.ToString(user.UserName)]
			var item *User = &User{
				User:                user,
				Groups:              iamClient.listGroupsForUser(aws.ToString(user.UserName)),
				AccessKeys:          iamClient.listAccessKeys(aws.ToString(user.UserName)),
				LoginProfile:        iamClient.listLoginProfile(aws.ToString(user.UserName)),
				InlinePolicies:      iamClient.listInlinePolicies(aws.ToString(user.UserName), "user"),
				AttachedPolicies:    iamClient.listAttachedPolicies(aws.ToString(user.UserName), "user"),
				PasswordEnabled:     userAccount.PasswordEnabled,
				PasswordLastChanged: userAccount.PasswordLastChanged,
				MfaActive:           userAccount.MfaActive,
			}
			mu.Lock()
			users = append(users, item)
		}(user)
	}
	wg.Wait()

	sort.Slice(users, func(i, j int) bool {
		return aws.ToString(users[i].UserName) < aws.ToString(users[j].UserName)
	})

	return
}

func listUsers() []types.User {
	var (
		marker         *string
		collectedUsers []types.User
	)

	for {
		output, err := iamClient.client.ListUsers(context.TODO(), &iam.ListUsersInput{
			Marker: marker,
		})
		if errors.As(err, &re) {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - users", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
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
func GetCredentialReport(cfg awsconfig.AWSConfig) (credentialReport map[string]*CredentialReport) {
	var (
		countRetries = 0
		maxRetries   = 5
	)

	iamClient = IAMClient{Config: cfg, client: iam.NewFromConfig(cfg.Config)}
	output, err := iamClient.client.GetCredentialReport(context.TODO(), &iam.GetCredentialReportInput{})

	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 410 { // Gone: https://http.cat/410
			checkGen, err := iamClient.client.GenerateCredentialReport(context.TODO(), &iam.GenerateCredentialReportInput{})
			if errors.As(err, &re) {
				log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - credential report", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
			}
			log.Println("Credential Report generation requested...")
			for checkGen.State != "COMPLETE" {
				if countRetries >= maxRetries {
					log.Println("Failed to generate Credential Report!")
					log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
				}
				countRetries++
				time.Sleep(5 * time.Second)
				checkGen, err = iamClient.client.GenerateCredentialReport(context.TODO(), &iam.GenerateCredentialReportInput{})
				if errors.As(err, &re) {
					log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
				}
			}
			return GetCredentialReport(cfg)
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - credential report", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	credentialReportCSV := []*CredentialReport{}
	if err := gocsv.Unmarshal(bytes.NewReader(output.Content), &credentialReportCSV); err != nil {
		log.Fatal(err)
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
		log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - access keys", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}
	accessKeys = output.AccessKeyMetadata
	return
}

func (ic *IAMClient) listLoginProfile(identity string) (loginProfile types.LoginProfile) {
	output, err := ic.client.GetLoginProfile(context.TODO(), &iam.GetLoginProfileInput{
		UserName: &identity,
	})
	if errors.As(err, &re) {
		// If 404 no login profile is configured
		if !(re.HTTPStatusCode() == 404) {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - login profiles", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	} else {
		loginProfile = *output.LoginProfile
	}
	return
}
