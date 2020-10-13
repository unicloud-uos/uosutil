package lib

import (
	"github.com/journeymidnight/aws-sdk-go/aws"
	"github.com/journeymidnight/aws-sdk-go/aws/credentials"
	"github.com/journeymidnight/aws-sdk-go/aws/endpoints"
	"github.com/journeymidnight/aws-sdk-go/aws/session"
	aws3 "github.com/journeymidnight/aws-sdk-go/service/s3"
	uos "github.com/unicloud-uos/uos-sdk-go/s3"
	"github.com/unicloud-uos/uos-sdk-go/s3/credential"
	"github.com/unicloud-uos/uos-sdk-go/s3/helper"
	"github.com/unicloud-uos/uos-sdk-go/s3/log"
	"github.com/uosutil/config"
	"strings"
)

type S3Client struct {
	AwsClient *aws3.S3
	UosClient *uos.Client
}

func NewClient(config *config.Config, bucketName string) (*S3Client, error) {
	if strings.ToLower(config.UseSDK) == "uos" {
		cfg := helper.GetDefaultConfig()
		if config.AccessKey != "" && config.SecretKey != "" {
			cfg.Credentials = credential.NewStaticCredentials(config.AccessKey, config.SecretKey, "")
		}
		if config.HostBucket == "" {
			if config.HostBase != "" {
				cfg.Endpoint = config.HostBase
			}
		} else {
			host := strings.ReplaceAll(config.HostBucket, "%(bucket)s", bucketName)
			cfg.Endpoint = host
		}
		cfg.Logger = log.NewLogger("error")

		client := uos.NewClient(*cfg)
		return &S3Client{UosClient: client}, nil
	} else {
		sessionConfig := buildSessionConfig(config)

		if config.HostBucket == "" || config.HostBucket == "%(bucket)s.s3.amazonaws.com" {
			svc := SessionNew(config)

			if loc, err := svc.GetBucketLocation(&aws3.GetBucketLocationInput{Bucket: &bucketName}); err != nil {
				return nil, err
			} else if loc.LocationConstraint == nil {
				// Use default service
				return &S3Client{AwsClient: svc}, nil
			} else {
				sessionConfig.Region = loc.LocationConstraint
			}
		} else {
			host := strings.ReplaceAll(config.HostBucket, "%(bucket)s", bucketName)

			sessionConfig.EndpointResolver = buildEndpointResolver(host)
		}

		return &S3Client{AwsClient: aws3.New(session.Must(session.NewSessionWithOptions(session.Options{
			Config:            sessionConfig,
			SharedConfigState: session.SharedConfigEnable,
		})))}, nil
	}
}

// DefaultRegion to use for S3 credential creation
const defaultRegion = "us-east-1"

func buildSessionConfig(config *config.Config) aws.Config {
	// By default make sure a region is specified, this is required for S3 operations
	sessionConfig := aws.Config{Region: aws.String(defaultRegion)}

	if config.AccessKey != "" && config.SecretKey != "" {
		sessionConfig.Credentials = credentials.NewStaticCredentials(config.AccessKey, config.SecretKey, "")
	}

	return sessionConfig
}

func buildEndpointResolver(hostname string) endpoints.Resolver {
	defaultResolver := endpoints.DefaultResolver()

	fixedHost := hostname
	if !strings.HasPrefix(hostname, "http") {
		fixedHost = "https://" + hostname
	}

	return endpoints.ResolverFunc(func(service, region string, optFns ...func(*endpoints.Options)) (endpoints.ResolvedEndpoint, error) {
		if service == endpoints.S3ServiceID {
			return endpoints.ResolvedEndpoint{
				URL: fixedHost,
			}, nil
		}

		return defaultResolver.EndpointFor(service, region, optFns...)
	})
}

// SessionNew - Read the config for default credentials, if not provided use environment based variables
func SessionNew(config *config.Config) *aws3.S3 {
	sessionConfig := buildSessionConfig(config)

	if config.HostBase != "" && config.HostBase != "s3.amazon.com" {
		sessionConfig.EndpointResolver = buildEndpointResolver(config.HostBase)
	}

	return aws3.New(session.Must(session.NewSessionWithOptions(session.Options{
		Config:            sessionConfig,
		SharedConfigState: session.SharedConfigEnable,
	})))
}
