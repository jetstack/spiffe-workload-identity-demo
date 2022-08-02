package main

import (
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Usage:     "SVID to external credential helper",
		ArgsUsage: "",
		Commands:  nil,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:      "workload-api-socket",
				Aliases:   []string{"w"},
				Usage:     "Path to SPIFFE workload API socket",
				Required:  false,
				Hidden:    false,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tls-cert-file",
				Aliases:   []string{"cert"},
				Usage:     "Path to X509 SVID cert file",
				Required:  false,
				Hidden:    false,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "tls-key-file",
				Aliases:   []string{"key"},
				Usage:     "Path to X509 SVID private key file",
				Required:  false,
				Hidden:    false,
				TakesFile: true,
			},
			&cli.StringFlag{
				Name:      "trusted-ca-file",
				Aliases:   []string{"ca"},
				Usage:     "Path to CAs that are trusted to sign SVIDs",
				Required:  false,
				Hidden:    false,
				TakesFile: true,
			},
		},
		Action:                 Run,
		UseShortOptionHandling: false,
	}
	app.Run(os.Args)
}
