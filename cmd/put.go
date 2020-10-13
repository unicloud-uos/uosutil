package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/uosutil/config"
	"github.com/uosutil/lib"
	"github.com/uosutil/utils"
	"github.com/urfave/cli"
)

func PutFunc(config *config.Config,args cli.Args) error {
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
		if err := put(config, u, dstUri); err != nil {
			return err
		}
	}

	return nil
}

func put(config *config.Config, src, dst *utils.FileURI) error {
	if src.Scheme != "file" && src.Scheme != "s3" {
		return errors.New("cp only supports local and s3 URLs")
	}
	if dst.Scheme != "file" && dst.Scheme != "s3" {
		return errors.New("cp only supports local and s3 URLs")
	}

	if config.Recursive {
		// Get the local file list and start copying
		err := filepath.Walk(src.Path, func(path string, info os.FileInfo, _ error) error {
			if info == nil || info.IsDir() {
				return nil
			}

			dstPath := dst.String()
			if strings.HasSuffix(dst.String(), "/") {
				dstPath += path
			} else {
				dstPath += "/" + path
			}
			dstUri, _ := utils.FileURINew(dstPath)
			dstUri.Scheme = dst.Scheme
			srcUri, _ := utils.FileURINew("file://" + path)

			return lib.CopyFile(config, srcUri, dstUri, true)
		})
		if err != nil {
			return err
		}
	} else {
		return lib.CopyFile(config, src, dst, false)
	}
	return nil
}
