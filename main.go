package main

import (
	"fmt"
	"github.com/rancher/binoculars/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

const (
	FlagApplicationName = "application-name"
	EnvApplicationName  = "APPLICATION_NAME"
	FlagInfluxDBURL     = "influxdb-url"
	EnvInfluxDBURL      = "INFLUXDB_URL"
	FlagInfluxDBUser    = "influxdb-user"
	EnvInfluxDBUser     = "INFLUXDB_USER"
	FlagInfluxDBPass    = "influxdb-pass"
	EnvInfluxDBPass     = "INFLUXDB_PASS"
	FlagQueryPeriod     = "query-period"
	EnvQueryPeriod      = "QUERY_PERIOD"
	FlagPort            = "port"
	EnvPort             = "PORT"
)

func main() {
	app := cli.NewApp()
	app.Name = "binoculars"
	app.Version = version.FriendlyVersion()
	app.Usage = "testy needs help!"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug, d",
			Usage:  "enable debug logging level",
			EnvVar: "DEBUG",
		},
	}
	app.Before = func(c *cli.Context) error {
		if c.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		return nil
	}

	app.Commands = []cli.Command{
		BinocularsCmd(),
	}

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func BinocularsCmd() cli.Command {
	return cli.Command{
		Name: "start",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   FlagApplicationName,
				EnvVar: EnvApplicationName,
				Usage:  "Specify the name of the application that is using this binoculars. This will be used to create a database name <application-name>_binoculars in the InfluxDB to store all data for this binoculars",
			},
		},
		Action: func(c *cli.Context) error {
			return startBinoculars(c)
		},
	}
}

func startBinoculars(c *cli.Context) error {
	if err := validateCommandLineArguments(c); err != nil {
		return err
	}
	logrus.Info("Hello word from binoculars!")

	applicationName := c.String(FlagApplicationName)
	logrus.Infof("FlagApplicationName: %v", applicationName)

	return nil
}

func validateCommandLineArguments(c *cli.Context) error {
	applicationName := c.String(FlagApplicationName)
	if applicationName == "" {
		return fmt.Errorf("no application name specified")
	}

	return nil
}
