package cmd

import (
	"github.com/uosutil/config"
	"github.com/urfave/cli"
)

func PutFunc(args cli.Args) error {
	file, err := config.LoadConfigFile()

	return nil
}
