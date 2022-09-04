package s3

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sort"
	"strings"
	"sync"

	awsconfig "nuvola/config/aws"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"golang.org/x/sync/semaphore"
)

func ListBuckets(cfg awsconfig.AWSConfig) (buckets []*Bucket) {
	var (
		s3Client = S3Client{Config: cfg, client: s3.NewFromConfig(cfg.Config)}
		mu       = &sync.Mutex{}
		sem      = semaphore.NewWeighted(int64(20)) // TODO: parametric
		wg       sync.WaitGroup
		re       *awshttp.ResponseError
	)

	output, err := s3Client.client.ListBuckets(context.TODO(), &s3.ListBucketsInput{})
	if errors.As(err, &re) {
		log.Fatalf("RequestID: %s, StatusCode: %d, error: %v\n", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
	}

	for _, bucket := range output.Buckets {
		wg.Add(1)
		go func(bucket types.Bucket) {
			if err := sem.Acquire(context.Background(), 1); err != nil {
				log.Fatalf("Failed to acquire semaphore: %v\n", err)
			}
			defer sem.Release(1)
			defer mu.Unlock()
			defer wg.Done()
			var item *Bucket = &Bucket{
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
	var (
		re  *awshttp.ResponseError
		err error
	)

	for _, region := range awsconfig.Regions {
		sc.Config.Config.Region = region
		sc.client = s3.NewFromConfig(sc.Config.Config)
		output, err = dumpFunction()
		if !errors.As(err, &re) {
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
			log.Fatalf("Service: %s, Function: %s, Error: %s\n", "S3", "getBucketPolicy", err)
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
			log.Fatalf("RequestID: %s, StatusCode: %d, error: %v", re.ServiceRequestID(), re.HTTPStatusCode(), re.Unwrap())
		}
	}
	return
}
