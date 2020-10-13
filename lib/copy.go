package lib

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/journeymidnight/aws-sdk-go/aws"
	"github.com/journeymidnight/aws-sdk-go/service/s3"
	"github.com/journeymidnight/aws-sdk-go/service/s3/s3manager"
	"github.com/uosutil/config"
	"github.com/uosutil/utils"
)

func CopyFile(config *config.Config, src, dst *utils.FileURI, ensureDirectory bool) error {
	if config.Verbose {
		fmt.Printf("Copy %s -> %s\n", src.String(), dst.String())
	}
	if config.DryRun {
		return nil
	}

	switch src.Scheme + "->" + dst.Scheme {
	case "file->file":
		return fmt.Errorf("cp should not be doing local files")
	case "s3->s3":
		return copyOnS3(config, src, dst)
	case "s3->file":
		return copyToLocal(config, src, dst, ensureDirectory)
	case "file->s3":
		return copyToS3(config, src, dst)
	}
	return nil
}

// Copy from S3 to local file
func copyToLocal(config *config.Config, src, dst *utils.FileURI, ensureDirectory bool) error {
	client, err := NewClient(config, src.Bucket)
	if err != nil {
		return err
	}
	downloader := s3manager.NewDownloaderWithClient(client.AwsClient)

	params := &s3.GetObjectInput{
		Bucket: aws.String(src.Bucket),
		Key:    src.Key(),
	}

	dst_path := dst.Path

	// if the destination is a directory then copy to a file in the directory
	sinfo, err := os.Stat(dst_path)
	if err == nil && sinfo.IsDir() {
		dst_path = path.Join(dst_path, filepath.Base(src.Path))
	}

	if ensureDirectory {
		dir := filepath.Dir(dst.Path)
		if _, err := os.Stat(dir); err != nil {
			if config.Verbose {
				fmt.Printf("Making directory dir=%s\n", dir)
			}
			if err := os.MkdirAll(dir, 0755); err != nil {
				fmt.Println(err)
				return fmt.Errorf("Error making directory dir=%s error=%v", dir, err)
			}
		}
	}

	fd, err := os.Create(dst_path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer fd.Close()

	_, err = downloader.Download(fd, params)
	if err != nil {
		return err
	}

	return nil
}

// Copy from local file to S3
func copyToS3(config *config.Config, src, dst *utils.FileURI) error {
	client, err := NewClient(config, dst.Bucket)
	if err != nil {
		return err
	}

	uploader := s3manager.NewUploaderWithClient(client.AwsClient, func(u *s3manager.Uploader) {
		u.PartSize = config.PartSize * 1024 * 1024
		u.Concurrency = config.Concurrency
	})

	fd, err := os.Open(src.Path)
	if err != nil {
		return err
	}
	defer fd.Close()

	params := &s3manager.UploadInput{
		Bucket: aws.String(dst.Bucket), // Required
		Key:    cleanBucketDestPath(src.Path, dst.Path),
		Body:   fd,
	}

	if config.StorageClass != "" {
		params.StorageClass = aws.String(config.StorageClass)
	}

	_, err = uploader.Upload(params)
	if err != nil {
		return err
	}

	return nil
}

// Copy from S3 to S3
//  -- if src and dst are the same it effects a "touch"
func copyOnS3(config *config.Config, src, dst *utils.FileURI) error {
	client, err := NewClient(config, dst.Bucket)
	if err != nil {
		return err
	}

	if strings.HasSuffix(src.Path, "/") {
		return fmt.Errorf("Invalid source for bucket to bucket copy path ends in '/'")
	}

	params := &s3.CopyObjectInput{
		Bucket:     aws.String(dst.Bucket),
		CopySource: aws.String(fmt.Sprintf("/%s/%s", src.Bucket, src.Path[1:])),
		Key:        cleanBucketDestPath(src.Path, dst.Path),
	}

	// if this is an overwrite - note that
	if src.Bucket == dst.Bucket && *params.CopySource == fmt.Sprintf("/%s/%s", dst.Bucket, *params.Key) {
		params.MetadataDirective = aws.String("REPLACE")
	}

	_, err = client.AwsClient.CopyObject(params)
	if err != nil {
		return err
	}

	return nil
}

// Take a src and dst and make a valid destination path for the bucket
//  if the dst ends in "/" add the basename of the source to the object
//  make sure the leading "/" is stripped off
func cleanBucketDestPath(src, dst string) *string {
	if strings.HasSuffix(dst, "/") {
		dst += filepath.Base(src)
	}
	if strings.HasPrefix(dst, "/") {
		dst = dst[1:]
	}
	return &dst
}
