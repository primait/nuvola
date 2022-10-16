package iam

import (
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	aat "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type IAMClient struct {
	client *iam.Client
	Config aws.Config
}

type AAClient struct {
	client *accessanalyzer.Client
	Config aws.Config
}

// Override SDK PolicyVersion type
type PolicyVersion struct {
	types.PolicyVersion
	Document PolicyDocument
}

// Override SDK AttachedPolicy type
type AttachedPolicies struct {
	types.AttachedPolicy
	Versions   []PolicyVersion
	Validation []aat.ValidatePolicyFinding `json:"Validation,omitempty"`
}

// Override SDK Role type
type Role struct {
	types.Role
	Description              string             `json:"Description"`
	AssumeRolePolicyDocument PolicyDocument     `json:"AssumeRolePolicyDocument,omitempty"`
	AssumableBy              []string           `json:"AssumableBy,omitempty"`
	AttachedPolicies         []AttachedPolicies `json:"AttachedPolicies,omitempty"`
	InlinePolicies           []PolicyDocument   `json:"InlinePolicies,omitempty"`
	InstanceProfileID        string             `json:"InstanceProfileId,omitempty"`
	InstanceProfileArn       string             `json:"InstanceProfileArn,omitempty"`
}

// Override SDK User type
type User struct {
	types.User
	PasswordEnabled     string                    `json:"PasswordEnabled,omitempty"`
	PasswordLastChanged string                    `json:"PasswordLastChanged,omitempty"`
	MfaActive           string                    `json:"MFAStatus,omitempty"`
	Groups              []types.Group             `json:"Groups,omitempty"`
	AccessKeys          []types.AccessKeyMetadata `json:"AccessKeys,omitempty"`
	LoginProfile        types.LoginProfile        `json:"LoginProfile,omitempty"`
	AttachedPolicies    []AttachedPolicies        `json:"AttachedPolicies,omitempty"`
	InlinePolicies      []PolicyDocument          `json:"InlinePolicies,omitempty"`
}

// Override SDK Group type
type Group struct {
	types.Group
	AttachedPolicies []AttachedPolicies `json:"AttachedPolicies,omitempty"`
	InlinePolicies   []PolicyDocument   `json:"InlinePolicies,omitempty"`
}

// Struct to the credential report CSV output
type CredentialReport struct {
	User                  string `csv:"user"`
	Arn                   string `csv:"arn"`
	UserCreation          string `csv:"user_creation_time"`
	PasswordEnabled       string `csv:"password_enabled"` // The value for the AWS account root user is always not_supported.
	PasswordLastUsed      string `csv:"password_last_used"`
	PasswordLastChanged   string `csv:"password_last_changed"`
	PasswordNextRotation  string `csv:"password_next_rotation"`
	MfaActive             string `csv:"mfa_active"`
	AccessKey1Active      string `csv:"access_key_1_active"`
	AccessKey1LastRotated string `csv:"access_key_1_last_rotated"` // TODO convert to Time
	AccessKey2Active      string `csv:"access_key_2_active"`
	AccessKey2LastRotated string `csv:"access_key_2_last_rotated"` // TODO convert to Time
	Cert1Active           string `csv:"cert_1_active"`
	Cert2Active           string `csv:"cert_2_active"`
}

type PolicyDocument struct {
	PolicyName string                      `json:"PolicyName,omitempty"`
	Version    string                      `json:"Version,omitempty"`
	Statement  []Statement                 `json:"Statement,omitempty"`
	ID         string                      `json:"Id,omitempty"`
	Validation []aat.ValidatePolicyFinding `json:"Validation,omitempty"`
}

type Principal struct {
	Service   interface{} `json:"Service,omitempty"`
	AWS       interface{} `json:"AWS,omitempty"`
	Federated interface{} `json:"Federated,omitempty"`
}

type Statement struct {
	Sid       string      `json:"Sid,omitempty"`
	Effect    string      `json:"Effect"`
	Principal *Principal  `json:"Principal,omitempty"`
	Action    interface{} `json:"Action"`
	Resource  interface{} `json:"Resource,omitempty"`
	Condition interface{} `json:"Condition,omitempty"`
}

var (
	re          *awshttp.ResponseError
	iamClient   IAMClient
	ActionsMap  map[string][]string
	ActionsList []string
)

func sortStringSlice(unsortedInterface interface{}) []string {
	rawInterface := unsortedInterface.([]interface{})
	sortedSlice := make([]string, len(rawInterface))
	for i, v := range rawInterface {
		sortedSlice[i] = v.(string)
	}
	sort.Strings(sortedSlice)
	return sortedSlice
}
