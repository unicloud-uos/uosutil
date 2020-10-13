package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/uosutil/cmd"
	"github.com/uosutil/config"
	"github.com/urfave/cli"
)

type Option struct {
	Recursive bool
	Force     bool
}

var Global Option

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
			Destination: &Global.Recursive,
		},
		cli.BoolFlag{
			Name:        "f,force",
			Usage:       "Force overwrite and other dangerous operations.",
			Destination: &Global.Force,
		},
	}

	app.Commands = cli.Commands{
		{
			Name:  "put",
			Usage: "Put file into bucket \n\t    uosutil put FILE [FILE...] s3://BUCKET[/PREFIX]",
			Action: func(c *cli.Context) error {
				cfg, err := config.NewConfig()
				if err != nil {
					fmt.Println("Put file err: ", err)
					return err
				}
				if len(c.Args()) < 2 {
					cli.ShowCommandHelp(cli.NewContext(app, nil, nil), "put")
					return errors.New("Not enough parameters for command 'put'")
				}
				return cmd.PutFunc(cfg, c.Args())
			},
			Flags: app.Flags,
		},
		{
			Name:  "get",
			Usage: "Get file from bucket \n\t    uosutil get s3://BUCKET/OBJECT LOCAL_FILE",
			Action: func(c *cli.Context) error {
				cfg, err := config.NewConfig()
				if err != nil {
					fmt.Println("Put file err: ", err)
					return err
				}
				if len(c.Args()) < 2 {
					cli.ShowCommandHelp(cli.NewContext(app, nil, nil), "get")
					return errors.New("Not enough parameters for command 'get'")
				}
				return cmd.GetFunc(cfg, c.Args())
			},
			Flags: app.Flags,
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
