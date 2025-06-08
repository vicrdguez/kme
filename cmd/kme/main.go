package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

const (
	DB_DIR      = ".kobo/KoboReader.sqlite"
	MARK_DIR    = ".kobo/markups"
	TMP_IMG_DIR = "tempimg"
	OUT_DIR     = "./kme-out"
)

func main() {
	rootCmd := &cli.Command{
		Commands: []*cli.Command{
			extract(),
			list(),
		},
	}

	if err := rootCmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
