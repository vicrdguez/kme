package main

import (
	"context"
	"fmt"
	"kme/internal/bookmark"
	"kme/internal/convert"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/urfave/cli/v3"
)

func extract() *cli.Command {
	cmd := &cli.Command{
		Name:   "extract",
		Usage:  "Extract bookmarks from the Kobo device",
		Action: handleExtract,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "device",
				Usage:    "Location to the Kobo device",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "out",
				Usage: "Output directory for temporary images and final PDF",
				Value: OUT_DIR,
			},
			&cli.BoolFlag{ // By default images are deleted
				Name:  "keep",
				Usage: "Keep temporary images",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "select",
				Usage: "Select the book(s) to extract from, instead of doing all",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "markups",
				Usage: "Extract just markups",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "highlights",
				Usage: "Extract just highlights",
				Value: false,
			},
			&cli.BoolFlag{
				Name:  "copy",
				Usage: "Copy the Kobo DB and markups folder to a temporary location",
				Value: false,
			},
		},
	}

	return cmd
}

func handleExtract(ctx context.Context, cmd *cli.Command) error {
	device := cmd.String("device")
	dbPath := filepath.Join(device, DB_DIR)
	markPath := filepath.Join(device, MARK_DIR)
	out := cmd.String("out")
	keep := cmd.Bool("keep")
	sel := cmd.Bool("select")
	marks := cmd.Bool("markups")
	highs := cmd.Bool("highlights")
	// cpy := cmd.Bool("copy")

	if err := validate(device, dbPath, markPath, out, keep, sel); err != nil {
		return err
	}

	// if cpy {
	// 	ctime := time.Now().Local()
	// 	cpyDir := fmt.Sprintf("/tmp/%s_kobodevice", ctime.Format("2006-01-02"))
	// 	os.CopyFS(cpyDir, dbPath)
	// }

	if err := bookmark.ConnectKoboDB(dbPath); err != nil {
		cli.Exit(err, 1)
	}

	var books []string
	fmt.Println("Finding all Books with bookmarks...")
	books = bookmark.AllBooks()

	if sel {
		selection, err := fuzzyFind(books)
		if err != nil {
			cli.Exit(err, 1)
		}
		books = selection
	}

	if len(books) == 0 {
		cli.Exit("No books found or selected", 1)
	}

	fmt.Println("Finding all bookmarks...")
	bookmarks := bookmark.AllBookmarks(books)
	fmt.Println("Processing bookmarks...")

	for _, bm := range bookmarks {
		switch {
		case marks:
			processMarkups(bm, markPath, out, keep)
		case highs:
			processHighlights(bm, out)
		default:
			processMarkups(bm, markPath, out, keep)
			processHighlights(bm, out)
		}
	}

	return nil
}

func processMarkups(bm *bookmark.Bookmarks, markPath string, out string, keep bool) error {
	// If no switch used for markups / highlights we do both (like if there was an --all)
	wg := sync.WaitGroup{}
	bookOutDir := filepath.Join(out, bm.Book)
	fmt.Println("Extracting markups to ", bookOutDir)

	if err := os.Mkdir(bookOutDir, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("Error creating output directory for book %s: %s", bm.Book, err)
	}

	fmt.Println("Book: ", bm.Book, "Total bookmarks: ", len(bm.Markups), "generating images...")

	wg.Add(1)
	go func() error {
		for i, m := range bm.Marks() {
			fmt.Printf("\tBookmark [%d/%d] - Book: %s, \n", i, len(bm.Markups), bm.Book)

			if err := convert.OverlayMarkup(m, markPath, bookOutDir); err != nil {
				fmt.Println(err)
			}
		}
		wg.Done()
		return nil
	}()
	wg.Wait()

	bookmark.CloseKoboDB()

	fmt.Println("Generating PDF ...")
	err := convert.BuildPDF(bm, bookOutDir, keep)
	if err != nil {
		fmt.Errorf("Error generating PDF: %w", err)
	}
	return nil
}

func processHighlights(bm *bookmark.Bookmarks, out string) error {
	ctime := time.Now().Local()
	fname := fmt.Sprintf("%s-highlights.txt", ctime.Format("200601021504"))
	filePath := filepath.Join(out, bm.Book, fname)
	fmt.Println("Extracting highlights to ", filePath)

	file, err := os.Create(filePath)
	defer file.Close()

	if err != nil {
		return fmt.Errorf("Could not create highlights file for book %s: %s", bm.Book, err)
	}

	for _, h := range bm.Highs() {
		if _, err := file.WriteString(fmt.Sprintf("- %s\n", h.Format())); err != nil {
			return fmt.Errorf("Error writing to highlights file %s: %s", filePath, err)
		}
	}
	fmt.Println("All highlights extracted for book", bm.Book)
	return nil
}

func validate(
	device string,
	dbPath string,
	markPath string,
	out string,
	keep bool,
	sel bool,
) error {
	//validation

	if fi, err := os.Stat(device); os.IsNotExist(err) || !fi.IsDir() {
		return cli.Exit("Provided device does not exist or is not a directory", 1)
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return cli.Exit("provided Kobo database does not exist", 1)
	}

	if fi, err := os.Stat(markPath); os.IsNotExist(err) || !fi.IsDir() {
		return cli.Exit("provided markups directory does not exist or is not a directory: "+markPath, 1)
	}

	if err := os.Mkdir(out, 0755); err != nil && !os.IsExist(err) {
		return cli.Exit("Error creating output directory", 1)
	}

	return nil
}

func fuzzyFind(books []string) ([]string, error) {

	idx, err := fuzzyfinder.FindMulti(
		books,
		func(i int) string {
			return books[i]
		},
	)
	if err != nil {
		return nil, fmt.Errorf("Error with fuzzy finder: %w\n", err)
	}

	selected := []string{}
	for _, i := range idx {
		selected = append(selected, books[i])
	}

	if len(selected) > 0 {
		return selected, nil
	}
	return books, nil
}
