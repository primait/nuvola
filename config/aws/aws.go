package aws_config

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"nuvola/io/requests"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	aat "github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/itchyny/gojq"
	"github.com/ohler55/ojg/oj"
)

// Struct for Policy structured output instead of SDK url-encoded string
type PolicyDocument struct {
	PolicyName string                      `json:"PolicyName,omitempty"`
	Version    string                      `json:"Version,omitempty"`
	Statement  []Statement                 `json:"Statement,omitempty"`
	Id         string                      `json:"Id,omitempty"`
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

type AWSConfig struct {
	Profile string
	aws.Config
}

// This is far from perfect: only User, Group, Role and Policy is supported and action with multiple targets are simply "*"
var IAMActionResourceMap map[string]string = map[string]string{
	"AddClientIDToOpenIDConnectProvider":        "oidc-provider*",
	"AddRoleToInstanceProfile":                  "InstanceProfile",
	"AddUserToGroup":                            "Group",
	"AttachGroupPolicy":                         "Group",
	"AttachRolePolicy":                          "Role",
	"AttachUserPolicy":                          "User",
	"ChangePassword":                            "User",
	"CreateAccessKey":                           "User",
	"CreateAccountAlias":                        "",
	"CreateGroup":                               "",
	"CreateInstanceProfile":                     "",
	"CreateLoginProfile":                        "User",
	"CreateOpenIDConnectProvider":               "oidc-provider*",
	"CreatePolicy":                              "",
	"CreatePolicyVersion":                       "Policy",
	"CreateRole":                                "",
	"CreateSAMLProvider":                        "saml-provider*",
	"CreateServiceLinkedRole":                   "Role",
	"CreateServiceSpecificCredential":           "User",
	"CreateUser":                                "",
	"CreateVirtualMFADevice":                    "mfa*",
	"DeactivateMFADevice":                       "User",
	"DeleteAccessKey":                           "User",
	"DeleteAccountAlias":                        "",
	"DeleteAccountPasswordPolicy":               "",
	"DeleteGroup":                               "Group",
	"DeleteGroupPolicy":                         "Group",
	"DeleteInstanceProfile":                     "InstanceProfile",
	"DeleteLoginProfile":                        "User",
	"DeleteOpenIDConnectProvider":               "oidc-provider*",
	"DeletePolicy":                              "Policy",
	"DeletePolicyVersion":                       "Policy",
	"DeleteRolePermissionsBoundary":             "Role",
	"DeleteRolePolicy":                          "Role",
	"DeleteRole":                                "Role",
	"DeleteSAMLProvider":                        "saml-provider*",
	"DeleteServerCertificate":                   "server-certificate*",
	"DeleteServiceLinkedRole":                   "Role",
	"DeleteServiceSpecificCredential":           "User",
	"DeleteSigningCertificate":                  "User",
	"DeleteSSHPublicKey":                        "User",
	"DeleteUserPermissionsBoundary":             "User",
	"DeleteUserPolicy":                          "User",
	"DeleteUser":                                "User",
	"DeleteVirtualMFADevice":                    "mfa",
	"DetachGroupPolicy":                         "Group",
	"DetachRolePolicy":                          "Role",
	"DetachUserPolicy":                          "User",
	"EnableMFADevice":                           "User",
	"GenerateCredentialReport":                  "",
	"GenerateOrganizationsAccessReport":         "access-report*",
	"GenerateServiceLastAccessedDetails":        "*",
	"GetAccessKeyLastUsed":                      "User",
	"GetAccountAuthorizationDetails":            "",
	"GetAccountPasswordPolicy":                  "",
	"GetAccountSummary":                         "",
	"GetContextKeysForCustomPolicy":             "",
	"GetContextKeysForPrincipalPolicy":          "*",
	"GetCredentialReport":                       "",
	"GetGroup":                                  "Group",
	"GetGroupPolicy":                            "Group",
	"GetInstanceProfile":                        "InstanceProfile",
	"GetLoginProfile":                           "User",
	"GetOpenIDConnectProvider":                  "oidc-provider*",
	"GetOrganizationsAccessReport":              "",
	"GetPolicy":                                 "Policy",
	"GetPolicyVersion":                          "Policy",
	"GetRolePolicy":                             "Role",
	"GetRole":                                   "Role",
	"GetSAMLProvider":                           "saml-provider*",
	"GetServerCertificate":                      "server-certificate*",
	"GetServiceLastAccessedDetails":             "",
	"GetServiceLastAccessedDetailsWithEntities": "",
	"GetServiceLinkedRoleDeletionStatus":        "Role",
	"GetSSHPublicKey":                           "User",
	"GetUserPolicy":                             "User",
	"GetUser":                                   "User",
	"ListAccessKeys":                            "User",
	"ListAccountAliases":                        "",
	"ListAttachedGroupPolicies":                 "Group",
	"ListAttachedRolePolicies":                  "Role",
	"ListAttachedUserPolicies":                  "User",
	"ListEntitiesForPolicy":                     "Policy",
	"ListGroupPolicies":                         "Group",
	"ListGroups":                                "",
	"ListGroupsForUser":                         "User",
	"ListInstanceProfilesForRole":               "Role",
	"ListInstanceProfiles":                      "InstanceProfile",
	"ListInstanceProfileTags":                   "InstanceProfile",
	"ListMFADevices":                            "user",
	"ListMFADeviceTags":                         "mfa*",
	"ListOpenIDConnectProviders":                "",
	"ListOpenIDConnectProviderTags":             "oidc-provider*",
	"ListPolicies":                              "",
	"ListPoliciesGrantingServiceAccess":         "*",
	"ListPolicyTags":                            "Policy",
	"ListPolicyVersions":                        "Policy",
	"ListRolePolicies":                          "Role",
	"ListRoles":                                 "",
	"ListRoleTags":                              "Role",
	"ListSAMLProviders":                         "",
	"ListSAMLProviderTags":                      "saml-provider*",
	"ListServerCertificates":                    "",
	"ListServerCertificateTags":                 "server-certificate*",
	"ListServiceSpecificCredentials":            "User",
	"ListSigningCertificates":                   "User",
	"ListSSHPublicKeys":                         "User",
	"ListUserPolicies":                          "User",
	"ListUsers":                                 "",
	"ListUserTags":                              "User",
	"ListVirtualMFADevices":                     "",
	"PassRole":                                  "Role",
	"PutGroupPolicy":                            "Group",
	"PutRolePermissionsBoundary":                "Role",
	"PutRolePolicy":                             "Role",
	"PutUserPermissionsBoundary":                "User",
	"PutUserPolicy":                             "User",
	"RemoveClientIDFromOpenIDConnectProvider":   "oidc-provider*",
	"RemoveRoleFromInstanceProfile":             "InstanceProfile",
	"RemoveUserFromGroup":                       "Group",
	"ResetServiceSpecificCredential":            "User",
	"ResyncMFADevice":                           "User",
	"SetDefaultPolicyVersion":                   "Policy",
	"SetSecurityTokenServicePreferences":        "",
	"SimulateCustomPolicy":                      "",
	"SimulatePrincipalPolicy":                   "*",
	"TagInstanceProfile":                        "InstanceProfile",
	"TagMFADevice":                              "mfa*",
	"TagOpenIDConnectProvider":                  "oidc-provider*",
	"TagPolicy":                                 "Policy",
	"TagRole":                                   "Role",
	"TagSAMLProvider":                           "saml-provider*",
	"TagServerCertificate":                      "server-certificate*",
	"TagUser":                                   "User",
	"UntagInstanceProfile":                      "InstanceProfile",
	"UntagMFADevice":                            "mfa*",
	"UntagOpenIDConnectProvider":                "oidc-provider*",
	"UntagPolicy":                               "Policy",
	"UntagRole":                                 "Role",
	"UntagSAMLProvider":                         "saml-provider*",
	"UntagServerCertificate":                    "server-certificate*",
	"UntagUser":                                 "User",
	"UpdateAccessKey":                           "User",
	"UpdateAccountPasswordPolicy":               "",
	"UpdateAssumeRolePolicy":                    "Role",
	"UpdateGroup":                               "Group",
	"UpdateLoginProfile":                        "User",
	"UpdateOpenIDConnectProviderThumbprint":     "oidc-provider*",
	"UpdateRoleDescription":                     "Role",
	"UpdateRole":                                "Role",
	"UpdateSAMLProvider":                        "saml-provider*",
	"UpdateServerCertificate":                   "server-certificate*",
	"UpdateServiceSpecificCredential":           "User",
	"UpdateSigningCertificate":                  "User",
	"UpdateSSHPublicKey":                        "User",
	"UpdateUser":                                "User",
	"UploadServerCertificate":                   "server-certificate*",
	"UploadSigningCertificate":                  "User",
	"UploadSSHPublicKey":                        "User",
}

var (
	Regions      []string
	ActionsMap   map[string][]string
	ActionsList  []string // len(unique(ActionList)) ~= 13k
	Conditions   map[string]string
	countRetries int
)

func InitAWSConfiguration(profile string) (awsc AWSConfig) {
	// Load the Shared AWS Configuration (~/.aws/config)
	cfg, _ := config.LoadDefaultConfig(context.TODO(), config.WithSharedConfigProfile(profile))
	awsc = AWSConfig{Profile: profile, Config: cfg}
	SetActions()
	return
}

func (ac *AWSConfig) WaitAPILimit() {
	var (
		maxRetries = 100
		client     = sts.NewFromConfig(ac.Config)
		output     *sts.GetCallerIdentityOutput
		caller     = ""
	)

	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	if ok && details != nil {
		caller = details.Name()
	}

	for blocked := false; !blocked; blocked = len(*output.Account) > 0 {
		if countRetries >= maxRetries {
			log.Fatalln("AWS is blocking the requests!")
		}
		countRetries++
		log.Printf("%s: API rate limit triggered! Waiting...\n", caller)
		time.Sleep(time.Duration(countRetries+5) * time.Second)
		// Tests the API call: aws sts get-caller-identity
		output, _ = client.GetCallerIdentity(context.TODO(), &sts.GetCallerIdentityInput{})
	}

	countRetries = 0
}

func SetActions() {
	URL := "https://awspolicygen.s3.amazonaws.com/js/policies.js"
	resString := strings.Replace(requests.MakeGet(URL), "app.PolicyEditorConfig=", "", 1)

	obj, err := oj.ParseString(resString)
	if err != nil {
		log.Fatalln(err)
	}

	query, err := gojq.Parse(`.serviceMap[] | .StringPrefix as $prefix | .Actions[] | "\($prefix):\(.)"`)
	if err != nil {
		log.Fatalln(err)
	}

	iter := query.Run(obj)
	ActionsMap = make(map[string][]string, 0)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			log.Fatalln(err)
		}

		ActionsList = append(ActionsList, v.(string))
		split := strings.Split(v.(string), ":")
		ActionsMap[split[0]] = append(ActionsMap[split[0]], split[1])
	}

	ActionsList = unique(ActionsList)
}

func GetActionsStartingWith(fullAction string) (actions []string) {
	// if not contains and expansion, return it
	if !strings.Contains(fullAction, "*") {
		actions = append(actions, fullAction)
		return
	}

	fullAction = strings.ToLower(strings.ReplaceAll(fullAction, "*", ""))
	if len(fullAction) <= 0 {
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
