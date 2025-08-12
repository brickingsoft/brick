package cert

import (
	"context"

	"github.com/urfave/cli/v3"
)

const (
	Name  = "cert"
	Usage = "create a cert file"
)

var (
	Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "mode",
			Usage:    "generate mode, such as CA CERT",
			Required: true,
		},
		&cli.StringFlag{
			Name:  "type",
			Usage: "cert type, such as ECDSA RSA ED25519",
			Value: "ECDSA",
		},
		&cli.IntFlag{
			Name:  "expire",
			Usage: "expire days",
			Value: 30,
		},
		&cli.StringFlag{
			Name:  "cn",
			Usage: "common name",
			Value: "brick",
		},
		&cli.StringFlag{
			Name:  "ca_cert",
			Usage: "ca cert file path",
		},
		&cli.StringFlag{
			Name:  "ca_key",
			Usage: "ca key file path",
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "output dir path",
		},
	}
)

func Action(ctx context.Context, c *cli.Command) (err error) {

	return
}
