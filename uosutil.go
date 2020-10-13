package main

import (
	"errors"
	"log"
	"os"

	"github.com/uosutil/cmd"
	"github.com/urfave/cli"
)

type Option struct {
	Recursive bool
	Force     bool
}

var global Option

func main() {
	app := cli.NewApp()
	app.Name = "UosUtil"
	app.Usage = "A simple CLI util for access uos"
	app.Copyright = "Unicloud"
	app.Version = "0.0.1"
	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelp(cli.NewContext(app, nil, nil))
		return nil
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:        "r,recursive",
			Usage:       "Recursive upload, download or removal.",
			Destination: &global.Recursive,
		},
		cli.BoolFlag{
			Name:        "f,force",
			Usage:       "Force overwrite and other dangerous operations.",
			Destination: &global.Force,
		},
	}

	app.Commands = cli.Commands{
		{
			Name:  "put",
			Usage: "Put file into bucket \n\t    uosutil put FILE [FILE...] s3://BUCKET[/PREFIX]",
			Action: func(c *cli.Context) error {
				if len(c.Args()) <= 2 {
					cli.ShowCommandHelp(cli.NewContext(app, nil, nil), "set")
					return errors.New("Not enough parameters for command 'put'")
				}
				return cmd.PutFunc(c.Args())
			},
			ArgsUsage: "<key> <value>",
		},
		{
			Name:  "get",
			Usage: "Get file from bucket \n\t    uosutil get s3://BUCKET/OBJECT LOCAL_FILE",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 2 {
					cli.ShowCommandHelp(cli.NewContext(app, nil, nil), "set")
					return errors.New("Not enough parameters for command 'get'")
				}
				return cmd.GetFunc(c.Args())
			},
			ArgsUsage: "<key> <value>",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
