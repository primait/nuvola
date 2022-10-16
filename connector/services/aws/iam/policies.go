package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	nuvolaerror "nuvola/tools/error"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	aat "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

var VALIDATE = false

func (ic *IAMClient) ValidatePolicy(policy string) (findings []aat.ValidatePolicyFinding) {
	if !VALIDATE {
		return nil
	}
	aaClient := AAClient{Config: ic.Config, client: accessanalyzer.NewFromConfig(ic.Config)}

	output, err := aaClient.client.ValidatePolicy(context.TODO(), &accessanalyzer.ValidatePolicyInput{
		PolicyDocument: &policy,
		PolicyType:     aat.PolicyTypeIdentityPolicy,
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "IAM - Policies", "ValidatePolicy")
	}

	if output != nil && len(output.Findings) > 0 {
		findings = output.Findings
	}

	return
}

// aws iam list-{role,user}-policies
func (ic *IAMClient) listInlinePolicies(identity string, object string) []PolicyDocument {
	var (
		policyVersionDocument = PolicyDocument{}
		policies              []string
		decodedValue          string
		attached              []PolicyDocument
	)

	switch {
	case object == "role":
		var attachedPolicies *iam.ListRolePoliciesOutput
		attachedPolicies, err := ic.client.ListRolePolicies(context.TODO(), &iam.ListRolePoliciesInput{
			RoleName: &identity,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListRolePolicies")
		}
		policies = attachedPolicies.PolicyNames
	case object == "user":
		var attachedPolicies *iam.ListUserPoliciesOutput
		attachedPolicies, err := ic.client.ListUserPolicies(context.TODO(), &iam.ListUserPoliciesInput{
			UserName: &identity,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListUserPolicies")
		}
		policies = attachedPolicies.PolicyNames
	case object == "group":
		var attachedPolicies *iam.ListGroupPoliciesOutput
		attachedPolicies, err := ic.client.ListGroupPolicies(context.TODO(), &iam.ListGroupPoliciesInput{
			GroupName: &identity,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListGroupPolicies")
		}
		policies = attachedPolicies.PolicyNames
	default:
		nuvolaerror.HandleError(nil, "IAM - Policies", "FAILED: no user/role/group defined")
	}

	for i := range policies {
		switch {
		case object == "role":
			var inlinePolicy *iam.GetRolePolicyOutput
			inlinePolicy, err := ic.client.GetRolePolicy(context.TODO(), &iam.GetRolePolicyInput{
				PolicyName: &policies[i],
				RoleName:   &identity,
			})
			if errors.As(err, &re) {
				nuvolaerror.HandleAWSError(re, "IAM - Policies", "GetRolePolicy")
			}
			decodedValue, _ = url.QueryUnescape(*inlinePolicy.PolicyDocument)
		case object == "user":
			var inlinePolicy *iam.GetUserPolicyOutput
			inlinePolicy, err := ic.client.GetUserPolicy(context.TODO(), &iam.GetUserPolicyInput{
				PolicyName: &policies[i],
				UserName:   &identity,
			})
			if errors.As(err, &re) {
				nuvolaerror.HandleAWSError(re, "IAM - Policies", "GetUserPolicy")
			}
			decodedValue, _ = url.QueryUnescape(*inlinePolicy.PolicyDocument)
		case object == "group":
			var inlinePolicy *iam.GetGroupPolicyOutput
			inlinePolicy, err := ic.client.GetGroupPolicy(context.TODO(), &iam.GetGroupPolicyInput{
				PolicyName: &policies[i],
				GroupName:  &identity,
			})
			if errors.As(err, &re) {
				nuvolaerror.HandleAWSError(re, "IAM - Policies", "GetGroupPolicy")
			}
			decodedValue, _ = url.QueryUnescape(*inlinePolicy.PolicyDocument)
		default:
			nuvolaerror.HandleError(nil, "IAM - Policies", "FAILED: no user/role/group defined")
		}

		err := json.Unmarshal([]byte(decodedValue), &policyVersionDocument)
		if err != nil {
			nuvolaerror.HandleError(nil, "IAM - Policies", "Error on Unmarshalling policyVersionDocument")
		}
		policyVersionDocument.PolicyName = policies[i]
		policyVersionDocument.Validation = ic.ValidatePolicy(decodedValue)
		expandActions(&policyVersionDocument)
		attached = append(attached, policyVersionDocument)
	}

	return attached
}

// aws iam list-policy-versions
func (ic *IAMClient) listPolicyVersions(policyArn *string) (policyVersions []PolicyVersion) {
	var (
		policyVersionDocument = PolicyDocument{}
		maxItems              = int32(1)
	)

	versions, err := ic.client.ListPolicyVersions(context.TODO(), &iam.ListPolicyVersionsInput{
		PolicyArn: policyArn,
		MaxItems:  &maxItems,
	})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListPolicyVersions")
	}

	for _, policyVersion := range versions.Versions {
		pv, err := ic.client.GetPolicyVersion(context.TODO(), &iam.GetPolicyVersionInput{
			PolicyArn: policyArn,
			VersionId: policyVersion.VersionId,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "GetPolicyVersion")
		}
		decodedValue, _ := url.QueryUnescape(*pv.PolicyVersion.Document)
		err = json.Unmarshal([]byte(decodedValue), &policyVersionDocument)
		if err != nil {
			nuvolaerror.HandleError(err, "IAM - Policies", "Umarshalling policyVersionDocument")
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
	)

	switch {
	case object == "role":
		var attachedPolicies *iam.ListAttachedRolePoliciesOutput
		attachedPolicies, err := ic.client.ListAttachedRolePolicies(context.TODO(), &iam.ListAttachedRolePoliciesInput{
			RoleName: &identity,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListAttachedRolePolicies")
		}
		output = attachedPolicies.AttachedPolicies
	case object == "user":
		var attachedPolicies *iam.ListAttachedUserPoliciesOutput
		attachedPolicies, err := ic.client.ListAttachedUserPolicies(context.TODO(), &iam.ListAttachedUserPoliciesInput{
			UserName: &identity,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListAttachedUserPolicies")
		}
		output = attachedPolicies.AttachedPolicies
	case object == "group":
		var attachedPolicies *iam.ListAttachedGroupPoliciesOutput
		attachedPolicies, err := ic.client.ListAttachedGroupPolicies(context.TODO(), &iam.ListAttachedGroupPoliciesInput{
			GroupName: &identity,
		})
		if errors.As(err, &re) {
			nuvolaerror.HandleAWSError(re, "IAM - Policies", "ListAttachedGroupPolicies")
		}
		output = attachedPolicies.AttachedPolicies
	default:
		nuvolaerror.HandleError(nil, "IAM - Policies", "FAILED: no user/role/group defined")
	}

	for _, policy := range output {
		policyVersions := ic.listPolicyVersions(policy.PolicyArn)
		policyDocument, errj := json.Marshal(policyVersions[0].Document)
		if errj != nil {
			nuvolaerror.HandleError(errj, "IAM - Policies", "Umarshalling policyVersions[0].Document")
		}

		expandActions(&policyVersions[0].Document)
		findings := ic.ValidatePolicy(string(policyDocument))
		attached = append(attached, AttachedPolicies{
			AttachedPolicy: policy,
			Versions:       policyVersions,
			Validation:     findings,
		})
	}

	return
}

func expandActions(policy *PolicyDocument) {
	for i, statement := range policy.Statement {
		var realAction []string

		switch v := statement.Action.(type) {
		case []interface{}:
			// list of Actions
			for _, action := range statement.Action.([]interface{}) {
				realAction = append(realAction, getActionsStartingWith(action.(string))...)
			}
		case string:
			// single Action
			realAction = append(realAction, getActionsStartingWith(v)...)
		default:
			nuvolaerror.HandleError(nil, "IAM - Policies", fmt.Sprintf("expandActions: %v %v\n", policy, v))
		}

		// Update the struct
		policy.Statement[i].Action = unique(realAction)
	}
}

func getActionsStartingWith(fullAction string) (actions []string) {
	// if not contains and expansion, return it
	if !strings.Contains(fullAction, "*") {
		actions = append(actions, fullAction)
		return
	}

	fullAction = strings.ToLower(strings.ReplaceAll(fullAction, "*", ""))
	if len(fullAction) == 0 {
		return ActionsList
	}

	splitStr := strings.Split(fullAction, ":")
	service := strings.TrimLeft(splitStr[0], " *")
	action := splitStr[1]
	for _, v := range ActionsMap[service] {
		if strings.HasPrefix(strings.ToLower(v), action) {
			actions = append(actions, fmt.Sprintf("%s:%s", service, v))
		}
	}

	return unique(actions)
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
