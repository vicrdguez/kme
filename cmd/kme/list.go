package main

import (
	"context"
	"fmt"
	"kme/internal/bookmark"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

func list() *cli.Command {
	return &cli.Command{
		Name:   "list-books",
		Usage:  "List books with bookmarks",
		Action: handleList,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "device",
				Required: true,
				Action: validateDevice(true),
			},
		},
	}
}

func handleList(ctx context.Context, cmd *cli.Command) error {
	device := cmd.String("device")
	dbPath := filepath.Join(device, DB_DIR)

	if err := bookmark.ConnectKoboDB(dbPath); err != nil {
		cli.Exit(err, 1)
	}

	books := bookmark.AllBooks()
	fmt.Printf("Found %d books:\n", len(books))
	for _, b := range books {
		fmt.Println("\t- ", b)
	}

	return nil

}

func validateDevice(isList bool) func(context.Context, *cli.Command, string) error {
	return func(ctx context.Context, cmd *cli.Command, device string) error {
		dbPath := filepath.Join(device, DB_DIR)
		markPath := filepath.Join(device, MARK_DIR)

		if fi, err := os.Stat(device); os.IsNotExist(err) || !fi.IsDir() {
			return cli.Exit("Provided device does not exist or is not a directory", 1)
		}
		if _, err := os.Stat(dbPath); os.IsNotExist(err) {
			return cli.Exit("provided Kobo database does not exist", 1)
		}

		if fi, err := os.Stat(markPath); (os.IsNotExist(err) && !isList) || !fi.IsDir() {
			return cli.Exit("provided markups directory does not exist or is not a directory", 1)
		}

		return nil
	}
}
