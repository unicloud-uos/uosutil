package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/journeymidnight/aws-sdk-go/service/s3"
	"github.com/uosutil/config"
	"github.com/uosutil/lib"
	"github.com/uosutil/utils"
	"github.com/urfave/cli"
)

func GetFunc(config *config.Config, args cli.Args) error {
	dst, args := args[len(args)-1], args[:len(args)-1]

	dstUri, err := utils.FileURINew(dst)
	if err != nil {
		return errors.New("Invalid destination argument")
	}
	if dstUri.Scheme == "" {
		dstUri.Scheme = "file"
	}
	if dstUri.Path == "" {
		dstUri.Path = "/"
	}

	for _, path := range args {
		u, err := utils.FileURINew(path)
		if err != nil {
			return fmt.Errorf("Invalid destination argument")
		}
		if u.Scheme == "" {
			u.Scheme = "file"
		}
		if err := get(config, u, dstUri); err != nil {
			return err
		}
	}

	return nil
}

func get(config *config.Config, src, dst *utils.FileURI) error {
	if src.Scheme != "file" && src.Scheme != "s3" {
		return errors.New("cp only supports local and s3 URLs")
	}
	if dst.Scheme != "file" && dst.Scheme != "s3" {
		return errors.New("cp only supports local and s3 URLs")
	}

	if config.Recursive {
		// Get the remote file list and start copying
		client, err := lib.NewClient(config, src.Bucket)
		if err != nil {
			return err
		}

		// For recusive we should assume that the src path ends in '/' since it's a directory
		nsrc := src
		if !strings.HasSuffix(src.Path, "/") {
			nsrc = src.SetPath(src.Path + "/")
		}

		basePath := nsrc.Path

		remotePager(config, client.AwsClient, nsrc.String(), false, func(page *s3.ListObjectsV2Output) {
			for _, obj := range page.Contents {
				src_path := *obj.Key
				fmt.Printf("src_path=%s  basePath=%s\n", src_path, basePath)
				src_path = src_path[len(basePath):]

				fmt.Printf("new src_path = %s\n", src_path)

				// uri := fmt.Sprintf("/%s", src.Host, *obj.Key)
				dst_path := dst.String()
				if strings.HasSuffix(dst.String(), "/") {
					dst_path += src_path
				} else {
					dst_path += "/" + src_path
				}

				dst_uri, _ := utils.FileURINew(dst_path)
				dst_uri.Scheme = dst.Scheme
				src_uri, _ := utils.FileURINew("s3://" + src.Bucket + "/" + *obj.Key)

				lib.CopyFile(config, src_uri, dst_uri, true)
			}
		})
		if err != nil {
			return err
		}
	} else {
		return lib.CopyFile(config, src, dst, false)
	}
	return nil
}
