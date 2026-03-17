package storage

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// loadAWSConfig loads AWS configuration with support for Localstack
func loadAWSConfig() (aws.Config, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	var loadOpts []func(*awsconfig.LoadOptions) error
	loadOpts = append(loadOpts, awsconfig.WithRegion(region))
	if endpoint != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{URL: endpoint, HostnameImmutable: true}, nil
		})
		loadOpts = append(loadOpts, awsconfig.WithEndpointResolverWithOptions(customResolver))
	}

	return awsconfig.LoadDefaultConfig(context.Background(), loadOpts...)
}

type Signer struct {
	client *s3.PresignClient
}

func NewSignerFromEnv() (*Signer, error) {
	cfg, err := loadAWSConfig()
	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		endpoint := os.Getenv("S3_ENDPOINT")
		if endpoint != "" {
			o.UsePathStyle = true
		}
	})
	pre := s3.NewPresignClient(s3Client)
	return &Signer{client: pre}, nil
}

func (s *Signer) PresignGetObject(ctx context.Context, bucket, key string, expires time.Duration) (string, error) {
	out, err := s.client.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, func(po *s3.PresignOptions) {
		po.Expires = expires
	})
	if err != nil {
		return "", err
	}
	return out.URL, nil
}
