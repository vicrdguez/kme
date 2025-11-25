package bookmark

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// Intermediate representation of bookmarks from DB
type koboBookmark struct {
	id        sql.NullString
	bookTitle sql.NullString
	section   sql.NullString
	part      sql.NullString
	location  sql.NullString
	kind      sql.NullString
	text      sql.NullString
	color     sql.NullInt64
}

type KoboDB struct {
	db *sql.DB
}

var kdb *KoboDB

func ConnectKoboDB(dbPath string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("Could not connect to the Kobo database")
	}
	kdb = &KoboDB{db: db}
	return nil
}

func CloseKoboDB() {
	kdb.db.Close()
}

// PDFs not supported for now
func (self *KoboDB) fetchBooksWithBookmark() ([]string, error) {
	var title sql.NullString
	query := `
	SELECT DISTINCT BookTitle
	FROM content
	INNER JOIN Bookmark ON content.BookID = Bookmark.VolumeID
	WHERE content.MimeType <> "application/pdf"
	`

	rows, err := self.db.Query(query)
	if err != nil {
		log.Fatalf("Error executing Books query: %v", err)
	}
	defer rows.Close()

	if err != nil {
		log.Fatalf("Could not fetch Books with bookmarks from Kobo database: %v", err)
		return []string{}, err
	}

	books := make([]string, 0, 50)

	for rows.Next() {
		err := rows.Scan(&title)
		if err != nil {
			continue
		}
		if title.Valid {
			books = append(books, title.String)
		}
	}
	return books, nil
}

func (self *KoboDB) fetchBookmarks(book string) []koboBookmark {
	// adobe_location is relevant only to PDFs, not supported for now
	// TODO Add PDF support
	query := `
	SELECT BookmarkID, BookTitle, Title, StartContainerPath, Type, Text, Color
	FROM content
	INNER JOIN Bookmark ON content.BookID = Bookmark.VolumeID
		AND content.ContentID = bookmark.ContentID
	WHERE BookTitle = ?1
	`
	stmt, err := self.db.Prepare(query)
	if err != nil {
		log.Fatalf("Error preparing Bookmarks query: %v", err)
	}

	rows, err := stmt.Query(book)
	if err != nil {
		log.Fatalf("Error executing Bookmarks query: %v", err)
	}
	defer rows.Close()

	if err != nil {
		log.Fatalf("Error finding Bookmarks: %v", err)
	}

	return fromRows(rows)
}

func fromRows(rows *sql.Rows) []koboBookmark {
	var bmList []koboBookmark
	for rows.Next() {
		bm := koboBookmark{}
		if err := rows.Scan(
			&bm.id,
			&bm.bookTitle,
			&bm.section,
			&bm.location,
			&bm.kind,
			&bm.text,
			&bm.color,
		); err != nil {
			log.Fatalf("Could not extract Bookmark info from DB: %v", err)
		}
		bmList = append(bmList, bm)
	}

	return bmList
}
