package s3

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/primait/nuvola/pkg/connector/services/aws/ec2"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/sourcegraph/conc/iter"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func ListBuckets(cfg aws.Config) (buckets []*Bucket) {
	s3Client := S3Client{Config: cfg, client: s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})}

	output, err := s3Client.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if errors.As(err, &re) {
		logging.HandleAWSError(re, "S3", "ListBuckets")
	}

	buckets = iter.Map(output.Buckets, func(bucket *types.Bucket) *Bucket {
		return &Bucket{
			Bucket:    *bucket,
			Policy:    s3Client.getBucketPolicy(bucket.Name),
			ACL:       s3Client.listBucketACL(bucket.Name),
			Encrypted: s3Client.getEncryptionStatus(bucket.Name),
		}
	})

	sort.Slice(buckets, func(i, j int) bool {
		return aws.ToString(buckets[i].Name) < aws.ToString(buckets[j].Name)
	})

	return
}

func (sc *S3Client) loopRegions(dumpFunction func() (interface{}, error)) (output interface{}) {
	var err error
	for _, region := range ec2.Regions {
		sc.Config.Region = region
		sc.client = s3.NewFromConfig(sc.Config)
		output, err = dumpFunction()
		if errors.As(err, &re) {
			return
		}
	}
	return
}

func (sc *S3Client) getBucketPolicy(bucket *string) (policy s3PolicyDocument) {
	var re *awshttp.ResponseError

	output, err := sc.client.GetBucketPolicy(context.TODO(), &s3.GetBucketPolicyInput{
		Bucket: bucket,
	})

	retry := func() (out interface{}) {
		return sc.loopRegions(func() (interface{}, error) {
			return sc.client.GetBucketPolicy(context.TODO(), &s3.GetBucketPolicyInput{
				Bucket: bucket,
			})
		})
	}

	if errors.As(err, &re) {
		if out := handleErrors(err, retry); out != nil {
			output = out.(*s3.GetBucketPolicyOutput)
		}
	}

	if output != nil {
		err := json.Unmarshal([]byte(aws.ToString(output.Policy)), &policy)
		if err != nil {
			logging.HandleError(err, "S3", "getBucketPolicy")
		}
	}
	return
}

func (sc *S3Client) listBucketACL(bucket *string) (grants []types.Grant) {
	var (
		output *s3.GetBucketAclOutput
		re     *awshttp.ResponseError
		err    error
	)

	output, err = sc.client.GetBucketAcl(context.TODO(), &s3.GetBucketAclInput{
		Bucket: bucket,
	})

	retry := func() (out interface{}) {
		return sc.loopRegions(func() (interface{}, error) {
			return sc.client.GetBucketAcl(context.TODO(), &s3.GetBucketAclInput{
				Bucket: bucket,
			})
		})
	}

	if errors.As(err, &re) {
		if out := handleErrors(err, retry); out != nil {
			output = out.(*s3.GetBucketAclOutput)
		}
	}

	if output != nil {
		grants = output.Grants
	}
	return
}

func (sc *S3Client) getEncryptionStatus(bucket *string) bool {
	var (
		output *s3.GetBucketEncryptionOutput
		re     *awshttp.ResponseError
		err    error
	)

	output, err = sc.client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
		Bucket: bucket,
	})

	retry := func() (out interface{}) {
		return sc.loopRegions(func() (interface{}, error) {
			return sc.client.GetBucketEncryption(context.TODO(), &s3.GetBucketEncryptionInput{
				Bucket: bucket,
			})
		})
	}

	if errors.As(err, &re) {
		if out := handleErrors(err, retry); out != nil {
			output = out.(*s3.GetBucketEncryptionOutput)
		}
	}

	return output != nil
}

func handleErrors(err error, retry func() interface{}) (output interface{}) {
	var re *awshttp.ResponseError

	if errors.As(err, &re) {
		switch re.HTTPStatusCode() {
		case 301:
			if strings.Contains(re.Unwrap().Error(), "PermanentRedirect") {
				return retry()
			}
		case 404, 0:
			// no policy applied to bucket, it's not illegal
			return
		case 400:
			if strings.Contains(re.Unwrap().Error(), "IllegalLocationConstraintException") {
				return retry()
			}
		default:
			logging.HandleAWSError(re, "S3", "handleErrors")
		}
	}
	return
}
