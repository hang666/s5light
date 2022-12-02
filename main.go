package main

import (
	"log"
	"os"

	s5light "github.com/hang666/s5light/server"
	"github.com/urfave/cli/v2"
)

func main() {
	var configPath string

	app := &cli.App{
		Name:    "s5light",
		Usage:   "A lightweight socks5 proxy server.",
		Version: "v0.1.0",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Value:       "",
				Usage:       "config file path",
				Destination: &configPath,
			},
		},
		Action: func(*cli.Context) error {
			s5light.SetConfigPath(configPath)
			s5light.ReadConfig()
			s5light.Server()
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}
