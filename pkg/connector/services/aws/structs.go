package awsconnector

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/primait/nuvola/pkg/io/logging"
)

// Struct for Policy structured output instead of SDK url-encoded string
// type PolicyDocument struct {
// 	PolicyName string                      `json:"PolicyName,omitempty"`
// 	Version    string                      `json:"Version,omitempty"`
// 	Statement  []Statement                 `json:"Statement,omitempty"`
// 	Id         string                      `json:"Id,omitempty"`
// 	Validation []aat.ValidatePolicyFinding `json:"Validation,omitempty"`
// }

// type Principal struct {
// 	Service   interface{} `json:"Service,omitempty"`
// 	AWS       interface{} `json:"AWS,omitempty"`
// 	Federated interface{} `json:"Federated,omitempty"`
// }

// type Statement struct {
// 	Sid       string      `json:"Sid,omitempty"`
// 	Effect    string      `json:"Effect"`
// 	Principal *Principal  `json:"Principal,omitempty"`
// 	Action    interface{} `json:"Action"`
// 	Resource  interface{} `json:"Resource,omitempty"`
// 	Condition interface{} `json:"Condition,omitempty"`
// }

type AWSConfig struct {
	Profile string
	aws.Config
	logger logging.LogManager
}

// This is far from perfect: only User, Group, Role and Policy is supported and action with multiple targets are simply "*"
var IAMActionResourceMap = map[string]string{
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
	"ListRoles":                                 "Role",
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
