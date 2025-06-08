package convert

import (
	"cmp"
	"fmt"
	"kme/internal/bookmark"
	"os"
	"path/filepath"
	"slices"
	"time"

	pdf "github.com/pdfcpu/pdfcpu/pkg/api"
)

func BuildPDF(bms *bookmark.Bookmarks, bookDir string, keep bool) error {

	// sorting ASC by loc, part, section
	slices.SortFunc(bms.Markups, func(a, b *bookmark.Markup) int {
		return cmp.Compare(a.OrderId, b.OrderId)
	})

	files := []string{}
	for _, bm := range bms.Marks() {
		fname := filepath.Join(bookDir, bm.Outfile())
		files = append(files, fname)
	}

	ctime := time.Now().Local()
	pdfFile := fmt.Sprintf(
		"%s - %s (markups).pdf",
		ctime.Format("20060102_1504"),
		bms.Book,
	)
	pdfOut := filepath.Join(bookDir, pdfFile)
	if err := pdf.ImportImagesFile(
		files,
		pdfOut,
		nil,
		nil,
	); err != nil {
		return err
	}

	fmt.Println("PDF saved: ", pdfOut)

	if !keep {
		fmt.Println("Deleting temporary images...")
		for _, f := range files {
			os.Remove(f)
		}
	}
	return nil
}
