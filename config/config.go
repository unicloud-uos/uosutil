package config

import (
	"errors"

	"github.com/go-ini/ini"
)

type Config struct {
	UseSDK       string `ini:"use_sdk"`
	AccessKey    string `ini:"access_key"`
	SecretKey    string `ini:"secret_key"`
	Recursive    bool   `ini:"recursive"`
	Force        bool   `ini:"force"`
	SkipExisting bool   `ini:"skip_existing"`
	HostBase     string `ini:"host_base"`
	HostBucket   string `ini:"host_bucket"`
	UseHttps     bool   `ini:"use_https"`
	StorageClass string `ini:"storage-class"`
	Concurrency  int    `ini:"concurrency"`
	PartSize     int64  `ini:"multipart_chunk_size_mb"`
	DryRun       bool   `ini:"dry_run"`
	Verbose      bool   `ini:"verbose"`
}

func NewConfig() (*Config, error) {
	config, err := loadConfigFile()
	if err != nil {
		return nil, err
	}
	return config, nil
}

func loadConfigFile() (*Config, error) {
	cfg := &Config{UseSDK: "AWS"}
	file, err := ini.Load("/root/.uoscfg")
	if err != nil {
		return cfg, nil
	}

	if _, err := file.Section("").NewKey("bucket", "%(bucket)s"); err != nil {
		return nil, errors.New("Unable to create bucket key")
	}

	if err := file.Section("default").MapTo(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
