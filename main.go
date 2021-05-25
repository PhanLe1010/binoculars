package main

import (
	"fmt"
	"github.com/rancher/binoculars/api"
	"github.com/rancher/binoculars/binoculars"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"

	"github.com/rancher/binoculars/pkg/version"
)

const (
	FlagApplicationName = "application-name"
	EnvApplicationName  = "APPLICATION_NAME"
	FlagDBURL           = "db-url"
	EnvDBURL            = "DB_URL"
	FlagDBUser          = "db-user"
	EnvDBUser           = "DB_USER"
	FlagDBPass          = "db-pass"
	EnvDBPass           = "DB_PASS"
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
			cli.StringFlag{
				Name:   FlagDBURL,
				EnvVar: EnvDBURL,
				Usage:  "Specify the URL of database",
			},
			cli.StringFlag{
				Name:   FlagDBUser,
				EnvVar: EnvDBUser,
				Usage:  "Specify the database user name",
			},
			cli.StringFlag{
				Name:   FlagDBPass,
				EnvVar: EnvDBPass,
				Usage:  "Specify the database password",
			},
			cli.StringFlag{
				Name:   FlagQueryPeriod,
				EnvVar: EnvQueryPeriod,
				Value:  "1h",
				Usage:  "Specify the period for how often each instance of the application makes the request. Cannot change after set for the first time. This value should be the same as time in GROUP BY clause in Grafana",
			},
			cli.IntFlag{
				Name:   FlagPort,
				EnvVar: EnvPort,
				Value:  8324,
				Usage:  "Specify the port number",
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

	applicationName := c.String(FlagApplicationName)
	dbURL := c.String(FlagDBURL)
	dbUser := c.String(FlagDBUser)
	dbPass := c.String(FlagDBPass)
	queryPeriod := c.String(FlagQueryPeriod)
	port := c.Int(FlagPort)

	done := make(chan struct{})

	server, err := binoculars.NewServer(done, applicationName, dbURL, dbUser, dbPass, queryPeriod)
	if err != nil {
		return err
	}
	router := http.Handler(api.NewRouter(server))

	listeningAddress := fmt.Sprintf("0.0.0.0:%v", port)

	go func() {
		logrus.Infof("Server is listening at %v", listeningAddress)
		// always returns error. ErrServerClosed on graceful close
		if err := http.ListenAndServe(listeningAddress, router); err != http.ErrServerClosed {
			logrus.Fatalf("%v", err)
		}
	}()

	RegisterShutdownChannel(done)
	<-done
	return nil
}

func RegisterShutdownChannel(done chan struct{}) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		logrus.Infof("Receive %v to exit", sig)
		close(done)
	}()
}

func validateCommandLineArguments(c *cli.Context) error {
	applicationName := c.String(FlagApplicationName)
	if applicationName == "" {
		return fmt.Errorf("no application name specified")
	}

	influxURL := c.String(FlagDBURL)
	if influxURL == "" {
		return fmt.Errorf("no database URL specified")
	}

	queryPeriod := c.String(FlagQueryPeriod)
	if _, err := time.ParseDuration(queryPeriod); err != nil {
		return errors.Wrap(err, "fail to parse --query-period")
	}

	return nil
}
