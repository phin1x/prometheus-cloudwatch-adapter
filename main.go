package main

import (
	"errors"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/urfave/cli/v2"
	"github.com/phin1x/prometheus-cloudwatch-adapter/pkg/adapter"
	"github.com/phin1x/prometheus-cloudwatch-adapter/pkg/logging"
)

func main() {
	logger := logging.GetLogger(false)

	app := &cli.App{
		Name:  "prometheus-cloudwatch-adapter",
		Usage: "forward prometheus metrics to cloudwatch",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "listen-address",
				Usage:   "Listen Address",
				Value:   ":9513",
				EnvVars: []string{"LISTEN_ADDRESS"},
			},

			&cli.StringFlag{
				Name:    "tls-cert",
				Usage:   "Path to tls cert",
				EnvVars: []string{"TLS_CERT"},
			},
			&cli.StringFlag{
				Name:    "tls-key",
				Usage:   "Path to tls key",
				EnvVars: []string{"TLS_KEY"},
			},

			&cli.StringFlag{
				Name:     "cloudwatch-namespace",
				Aliases:  []string{"n"},
				Required: true,
				Usage:    "AWS Cloudwatch Namespace",
				EnvVars:  []string{"CLOUDWATCH_NAMESPACE"},
			},

			&cli.BoolFlag{
				Name:    "debug",
				Usage:   "Enable debug messages",
				Value:   false,
				EnvVars: []string{"DEBUG"},
			},
		},
		Action: func(c *cli.Context) error {
			if _, found := os.LookupEnv("AWS_REGION"); !found {
				return errors.New("env var AWS_REGION not set")
			}

			a, err := adapter.New(&adapter.Configuration{
				ListenAddress:       c.String("listen-address"),
				TLSCert:             c.String("tls-cert"),
				TLSKey:              c.String("tls-key"),
				CloudwatchNamespace: c.String("cloudwatch-namespace"),
				Debug:               c.Bool("debug"),
			}, logger)
			if err != nil {
				return err
			}

			return a.Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Error(err, "an error occurred while running the app")
	}
}

func init() {
	prometheus.MustRegister(prometheus.NewBuildInfoCollector())
}
