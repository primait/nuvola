package s3

import (
	"context"
	"encoding/json"
	"errors"
	"nuvola/connector/services/aws/ec2"
	nuvolaerror "nuvola/tools/error"
	"sort"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/semaphore"
)

func ListBuckets(cfg aws.Config) (buckets []*Bucket) {
	var (
		s3Client = S3Client{Config: cfg, client: s3.NewFromConfig(cfg)}
		mu       = &sync.Mutex{}
		sem      = semaphore.NewWeighted(int64(15))
		wg       sync.WaitGroup
	)

	output, err := s3Client.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if errors.As(err, &re) {
		nuvolaerror.HandleAWSError(re, "S3", "ListBuckets")
	}

	for _, bucket := range output.Buckets {
		wg.Add(1)
		go func(bucket types.Bucket) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				nuvolaerror.HandleError(err, "S3", "ListBuckets - Acquire Semaphore")
			}
			defer sem.Release(1)
			defer mu.Unlock()
			defer wg.Done()
			var item = &Bucket{
				Bucket:    bucket,
				Policy:    s3Client.getBucketPolicy(bucket.Name),
				ACL:       s3Client.listBucketACL(bucket.Name),
				Encrypted: s3Client.getEncryptionStatus(bucket.Name),
			}
			mu.Lock()
			buckets = append(buckets, item)
		}(bucket)
	}
	wg.Wait()

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
			nuvolaerror.HandleError(err, "S3", "getBucketPolicy")
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
			nuvolaerror.HandleAWSError(re, "S3", "handleErrors")
		}
	}
	return
}
