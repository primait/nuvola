package iam

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/url"
	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	aat "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

var VALIDATE = false

func (ic *IAMClient) ValidatePolicy(policy string) (findings []aat.ValidatePolicyFinding) {
	if !VALIDATE {
		return findings
	}
	aaClient := AAClient{Config: ic.Config, client: accessanalyzer.NewFromConfig(ic.Config.Config)}

	output, err := aaClient.client.ValidatePolicy(context.TODO(), &accessanalyzer.ValidatePolicyInput{
		PolicyDocument: &policy,
		PolicyType:     aat.PolicyTypeIdentityPolicy,
	})
	if errors.As(err, &re) {
		if re.HTTPStatusCode() == 429 { // Too Many Requests https://http.cat/429
			ic.Config.WaitAPILimit()
			return ic.ValidatePolicy(policy)
		} else if re.HTTPStatusCode() == 400 {
			// deserialization failed, failed to decode response body
			return
		} else {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - validate policy", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}

	if output != nil && len(output.Findings) > 0 {
		findings = output.Findings
	}

	return
}

// aws iam list-{role,user}-policies
func (ic *IAMClient) listInlinePolicies(identity string, object string) (attached []awsconfig.PolicyDocument) {
	var (
		policyVersionDocument = awsconfig.PolicyDocument{}
		policies              []string
		decodedValue          string
		err                   error
	)

	switch {
	case object == "role":
		var attachedPolicies *iam.ListRolePoliciesOutput
		attachedPolicies, err = ic.client.ListRolePolicies(context.TODO(), &iam.ListRolePoliciesInput{
			RoleName: &identity,
		})
		policies = attachedPolicies.PolicyNames
	case object == "user":
		var attachedPolicies *iam.ListUserPoliciesOutput
		attachedPolicies, err = ic.client.ListUserPolicies(context.TODO(), &iam.ListUserPoliciesInput{
			UserName: &identity,
		})
		policies = attachedPolicies.PolicyNames
	case object == "group":
		var attachedPolicies *iam.ListGroupPoliciesOutput
		attachedPolicies, err = ic.client.ListGroupPolicies(context.TODO(), &iam.ListGroupPoliciesInput{
			GroupName: &identity,
		})
		policies = attachedPolicies.PolicyNames
	default:
		log.Fatalln("FAILED: no user/role/group defined")
	}

	if errors.As(err, &re) {
		log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - inline policies", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}

	for i := range policies {
		switch {
		case object == "role":
			var inlinePolicy *iam.GetRolePolicyOutput
			inlinePolicy, err = ic.client.GetRolePolicy(context.TODO(), &iam.GetRolePolicyInput{
				PolicyName: &policies[i],
				RoleName:   &identity,
			})
			decodedValue, _ = url.QueryUnescape(*inlinePolicy.PolicyDocument)
		case object == "user":
			var inlinePolicy *iam.GetUserPolicyOutput
			inlinePolicy, err = ic.client.GetUserPolicy(context.TODO(), &iam.GetUserPolicyInput{
				PolicyName: &policies[i],
				UserName:   &identity,
			})
			decodedValue, _ = url.QueryUnescape(*inlinePolicy.PolicyDocument)
		case object == "group":
			var inlinePolicy *iam.GetGroupPolicyOutput
			inlinePolicy, err = ic.client.GetGroupPolicy(context.TODO(), &iam.GetGroupPolicyInput{
				PolicyName: &policies[i],
				GroupName:  &identity,
			})
			decodedValue, _ = url.QueryUnescape(*inlinePolicy.PolicyDocument)
		default:
			log.Fatalln("FAILED: no user/role/group defined")
		}

		if errors.As(err, &re) {
			log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - inline policies", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}

		err := json.Unmarshal([]byte(decodedValue), &policyVersionDocument)
		if err != nil {
			log.Fatalln(err)
		}
		policyVersionDocument.PolicyName = policies[i]
		policyVersionDocument.Validation = ic.ValidatePolicy(decodedValue)
		expandActions(&policyVersionDocument)
		attached = append(attached, policyVersionDocument)
	}

	return
}

func expandActions(policy *awsconfig.PolicyDocument) {
	for i, statement := range policy.Statement {
		var realAction []string

		switch v := statement.Action.(type) {
		case []interface{}:
			// list of Actions
			for _, action := range statement.Action.([]interface{}) {
				realAction = append(realAction, awsconfig.GetActionsStartingWith(action.(string))...)
			}
		case string:
			// single Action
			realAction = append(realAction, awsconfig.GetActionsStartingWith(v)...)
		default:
			log.Fatalf("Case not implemented: %v\n", v)
		}

		// Update the struct
		policy.Statement[i].Action = unique(realAction)
	}
}

func unique(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

// aws iam list-policy-versions
func (ic *IAMClient) listPolicyVersions(policyArn *string) (policyVersions []PolicyVersion) {
	var (
		policyVersionDocument = awsconfig.PolicyDocument{}
		maxItems              = int32(1)
	)

	versions, err := ic.client.ListPolicyVersions(context.TODO(), &iam.ListPolicyVersionsInput{
		PolicyArn: policyArn,
		MaxItems:  &maxItems,
	})
	if errors.As(err, &re) {
		log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - policy version", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}

	for _, policyVersion := range versions.Versions {
		pv, err := ic.client.GetPolicyVersion(context.TODO(), &iam.GetPolicyVersionInput{
			PolicyArn: policyArn,
			VersionId: policyVersion.VersionId,
		})
		if err != nil {
			log.Fatal(err)
		}
		decodedValue, _ := url.QueryUnescape(*pv.PolicyVersion.Document)
		err = json.Unmarshal([]byte(decodedValue), &policyVersionDocument)
		if err != nil {
			log.Fatalln(err)
		}
		policyVersions = append(policyVersions, PolicyVersion{
			PolicyVersion: policyVersion,
			Document:      policyVersionDocument,
		})
	}

	return
}

// aws iam list-attached-{role,user}-policies
func (ic *IAMClient) listAttachedPolicies(identity string, object string) (attached []AttachedPolicies) {
	var (
		output []types.AttachedPolicy
		err    error
	)

	switch {
	case object == "role":
		var attachedPolicies *iam.ListAttachedRolePoliciesOutput
		attachedPolicies, err = ic.client.ListAttachedRolePolicies(context.TODO(), &iam.ListAttachedRolePoliciesInput{
			RoleName: &identity,
		})
		output = attachedPolicies.AttachedPolicies
	case object == "user":
		var attachedPolicies *iam.ListAttachedUserPoliciesOutput
		attachedPolicies, err = ic.client.ListAttachedUserPolicies(context.TODO(), &iam.ListAttachedUserPoliciesInput{
			UserName: &identity,
		})
		output = attachedPolicies.AttachedPolicies
	case object == "group":
		var attachedPolicies *iam.ListAttachedGroupPoliciesOutput
		attachedPolicies, err = ic.client.ListAttachedGroupPolicies(context.TODO(), &iam.ListAttachedGroupPoliciesInput{
			GroupName: &identity,
		})
		output = attachedPolicies.AttachedPolicies
	default:
		log.Fatalln("FAILED: no user/role/group defined")
	}

	if errors.As(err, &re) {
		log.Fatalf("Service: %s, RequestID: %s, StatusCode: %d, error: %v", "IAM - attached policies", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}

	for _, policy := range output {
		policyVersions := ic.listPolicyVersions(policy.PolicyArn)
		policyDocument, _ := json.Marshal(policyVersions[0].Document)
		expandActions(&policyVersions[0].Document)
		attached = append(attached, AttachedPolicies{
			AttachedPolicy: policy,
			Versions:       policyVersions,
			Validation:     ic.ValidatePolicy(string(policyDocument)),
		})
	}

	return
}
