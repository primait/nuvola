package lambda

import (
	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

type Lambda struct {
	types.FunctionConfiguration
	types.FunctionCodeLocation
	Policy lambdaPolicyDocument
}

type LambdaClient struct {
	client *lambda.Client
	Config awsconfig.AWSConfig
}

type statement struct {
	Sid       string      `json:"Sid,omitempty"`
	Effect    string      `json:"Effect"`
	Principal interface{} `json:"Principal,omitempty"`
	Action    interface{} `json:"Action"`
	Resource  interface{} `json:"Resource,omitempty"`
	Condition interface{} `json:"Condition,omitempty"`
}

type lambdaPolicyDocument struct {
	Version   string      `json:"Version,omitempty"`
	Id        string      `json:"Id,omitempty"`
	Statement []statement `json:"Statement,omitempty"`
	Condition interface{} `json:"Condition,omitempty"`
}
