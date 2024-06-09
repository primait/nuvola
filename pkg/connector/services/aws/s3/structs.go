package s3

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/primait/nuvola/pkg/io/logging"
)

type S3Client struct {
	client *s3.Client
	Config aws.Config
	logger logging.LogManager
}

// Override SDK S3 Bucket type
type Bucket struct {
	types.Bucket
	Policy    s3PolicyDocument `json:"Policy,omitempty"`
	ACL       []types.Grant    `json:"ACL,omitempty"`
	Encrypted bool
}

type statement struct {
	SID       string      `json:"Sid,omitempty"`
	Effect    string      `json:"Effect"`
	Principal interface{} `json:"Principal,omitempty"`
	Action    interface{} `json:"Action"`
	Resource  interface{} `json:"Resource,omitempty"`
	Condition interface{} `json:"Condition,omitempty"`
}

type s3PolicyDocument struct {
	Version   string      `json:"Version,omitempty"`
	ID        string      `json:"Id,omitempty"`
	Statement []statement `json:"Statement,omitempty"`
	Condition interface{} `json:"Condition,omitempty"`
}

var re *awshttp.ResponseError
