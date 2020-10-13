package cmd

import (
	"fmt"
	"github.com/journeymidnight/aws-sdk-go/aws"
	"github.com/journeymidnight/aws-sdk-go/service/s3"
	"github.com/uosutil/config"
	"github.com/uosutil/lib"
	"github.com/uosutil/utils"
)

func remotePager(config *config.Config, svc *s3.S3, uri string, delim bool, pager func(page *s3.ListObjectsV2Output)) error {
	u, err := utils.FileURINew(uri)
	if err != nil || u.Scheme != "s3" {
		return fmt.Errorf("requires buckets to be prefixed with s3://")
	}

	params := &s3.ListObjectsV2Input{
		Bucket:  aws.String(u.Bucket), // Required
		MaxKeys: aws.Int64(1000),
	}
	if u.Path != "" && u.Path != "/" {
		params.Prefix = u.Key()
	}
	if delim {
		params.Delimiter = aws.String("/")
	}

	wrapper := func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		pager(page)
		return true
	}

	if svc == nil {
		svc = lib.SessionNew(config)
	}

	bsvc, err := lib.NewClient(config, u.Bucket)
	if err != nil {
		return err
	}
	if err := bsvc.AwsClient.ListObjectsV2Pages(params, wrapper); err != nil {
		return err
	}
	return nil
}