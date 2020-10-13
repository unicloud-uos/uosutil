package utils

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
)

type FileURI struct {
	Scheme string
	Bucket string
	Path   string
}

func FileURINew(path string) (*FileURI, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "" && u.Scheme != "s3" && u.Scheme != "file" {
		return nil, errors.New("Invalid URI scheme, it must be one of file/s3/NONE")
	}

	uri := FileURI{
		Scheme: u.Scheme,
		Bucket: u.Host,
		Path:   u.Path,
	}

	if uri.Scheme == "" {
		uri.Scheme = "file"
	}
	if uri.Scheme == "s3" && uri.Path != "" {
		uri.Path = uri.Path[1:]
	}
	if uri.Path == "" && uri.Scheme == "s3" {
		uri.Path = "/"
	}

	return &uri, nil
}

func (uri *FileURI) Key() *string {
	if uri.Path[0] == '/' {
		s := uri.Path[1:]
		return &s
	}
	return &uri.Path
}

func (uri *FileURI) String() string {
	if uri.Scheme == "s3" {
		return fmt.Sprintf("s3://%s/%s", uri.Bucket, *uri.Key())
	} else {
		return fmt.Sprintf("file://%s", uri.Path)
	}
}

func (uri *FileURI) Join(elem string) *FileURI {
	nuri := FileURI{
		Scheme: uri.Scheme,
		Bucket: uri.Bucket,
	}

	if elem == "" {
		nuri.Path = uri.Path
	} else if elem[0] == '/' {
		nuri.Path = elem
	} else {
		// TODO: https://golang.org/pkg/net/url/#URL.ResolveReference
		nuri.Path = path.Join(filepath.Dir(uri.Path), elem)
		if elem[len(elem)-1] == '/' {
			nuri.Path += "/"
		}
	}

	return &nuri
}

func (uri *FileURI) SetPath(elem string) *FileURI {
	nuri := FileURI{
		Scheme: uri.Scheme,
		Bucket: uri.Bucket,
		Path:   elem,
	}
	if uri.Path == "" && uri.Scheme == "s3" {
		uri.Path = "/"
	}

	return &nuri
}
