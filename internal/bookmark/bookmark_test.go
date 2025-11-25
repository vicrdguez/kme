package bookmark

import (
	"cmp"
	"database/sql"
	"testing"
)

type parseTest struct {
	want       string
	numwant    float64
	in         sql.NullString
	shouldFail bool
}

func TestOrderId(t *testing.T) {
	tests := map[string]koboBookmark{
		// different chapter and location
		"big1": {
			id:        sql.NullString{String: "54930965-1037-4d2c-974d-2fcc05e6a274", Valid: true},
			bookTitle: sql.NullString{String: "Your mom", Valid: true},
			section:   sql.NullString{String: "xhtml/chapter3.xhtml", Valid: true},
			location:  sql.NullString{String: `span#kobo\.115\.1`, Valid: true},
			kind:      sql.NullString{String: `markup`, Valid: true},
			text:      sql.NullString{String: "", Valid: true},
		},
		"small1": {
			id:        sql.NullString{String: "54930965-1037-4d2c-974d-2fcc05e6a274", Valid: true},
			bookTitle: sql.NullString{String: "Your mom", Valid: true},
			section:   sql.NullString{String: "ch01.html", Valid: true},
			location:  sql.NullString{String: `span#kobo\.38\.3`, Valid: true},
			kind:      sql.NullString{String: `markup`, Valid: true},
			text:      sql.NullString{String: "", Valid: true},
		},
		// same chapter but different location
		"big2": {
			id:        sql.NullString{String: "54930965-1037-4d2c-974d-2fcc05e6a274", Valid: true},
			bookTitle: sql.NullString{String: "Your mom", Valid: true},
			section:   sql.NullString{String: "xhtml/chapter3.xhtml", Valid: true},
			location:  sql.NullString{String: `span#kobo\.115\.1`, Valid: true},
			kind:      sql.NullString{String: `markup`, Valid: true},
			text:      sql.NullString{String: "", Valid: true},
		},
		"small2": {
			id:        sql.NullString{String: "54930965-1037-4d2c-974d-2fcc05e6a274", Valid: true},
			bookTitle: sql.NullString{String: "Your mom", Valid: true},
			section:   sql.NullString{String: "xhtml/chapter3.xhtml", Valid: true},
			location:  sql.NullString{String: `span#kobo\.38\.3`, Valid: true},
			kind:      sql.NullString{String: `markup`, Valid: true},
			text:      sql.NullString{String: "", Valid: true},
		},
		"big3": {
			id:        sql.NullString{String: "efe64dca-64e1-4351-a6d9-dc475e7db003", Valid: true},
			bookTitle: sql.NullString{String: "Your mom", Valid: true},
			section:   sql.NullString{String: "xhtml/McKe_9780804137393_epub3_c019_r1.xhtml", Valid: true},
			location:  sql.NullString{String: `span#kobo.93.1`, Valid: true},
			kind:      sql.NullString{String: `markup`, Valid: true},
			text:      sql.NullString{String: "", Valid: true},
		},
		"small3": {
			id:        sql.NullString{String: "efe64dca-64e1-4351-a6d9-dc475e7db003", Valid: true},
			bookTitle: sql.NullString{String: "Your mom", Valid: true},
			section:   sql.NullString{String: "xhtml/McKe_9780804137393_epub3_c019_r1.xhtml", Valid: true},
			location:  sql.NullString{String: `span#kobo.22.8`, Valid: true},
			kind:      sql.NullString{String: `markup`, Valid: true},
			text:      sql.NullString{String: "", Valid: true},
		},
	}

	orderIds := map[string]float64{}

	for k, tt := range tests {
		bm := fromRawValues(tt).(*Markup)
		orderIds[k] = bm.OrderId
	}
	if cmp.Compare(orderIds["big1"], orderIds["small1"]) < 0 {
		t.Errorf(
			"Incorrect order ids. %f should be greater then %f",
			orderIds["big1"],
			orderIds["small1"],
		)
	}
	if cmp.Compare(orderIds["big2"], orderIds["small2"]) < 0 {
		t.Errorf(
			"Incorrect order ids. %f should be greater then %f",
			orderIds["big2"],
			orderIds["small2"],
		)
	}
	if cmp.Compare(orderIds["big3"], orderIds["small3"]) < 0 {
		t.Errorf(
			"Incorrect order ids. %f should be greater then %f",
			orderIds["big3"],
			orderIds["small3"],
		)
	}
}

func TestLocationProcessing(t *testing.T) {
	tests := []parseTest{
		{
			want:       "40.9",
			numwant:    0.0409,
			in:         sql.NullString{String: `span#kobo\.40\.9`, Valid: true},
			shouldFail: false,
		},
		{
			want:       "38.3",
			numwant:    0.0383,
			in:         sql.NullString{String: `span#kobo\.38\.3`, Valid: true},
			shouldFail: false,
		},
		{
			want:       "1.2",
			numwant:    0.0012,
			in:         sql.NullString{String: `span#kobo\.1\.2`, Valid: true},
			shouldFail: false,
		},
		{
			want:       "115.1",
			numwant:    0.1151,
			in:         sql.NullString{String: `span#kobo\.115\.1`, Valid: true},
			shouldFail: false,
		},
		{
			want:       "114.3",
			numwant:    0.1143,
			in:         sql.NullString{String: `span#kobo.114.3`, Valid: true},
			shouldFail: false,
		},
		{
			want:       "114.3",
			numwant:    0.1143,
			in:         sql.NullString{String: `span#kobo.114.3`, Valid: false},
			shouldFail: true,
		},
	}

	for _, tt := range tests {
		loc, numloc := parseLocation(tt.in)
		if (loc != tt.want || numloc != tt.numwant) && !tt.shouldFail {
			t.Errorf("Incorrect result: want: '%s', got: %s ; want: %v, got: %v", tt.want, loc, tt.numwant, numloc)
		}
	}
}

func TestParseSection(t *testing.T) {
	tests := []parseTest{
		{
			want:       "chapter03",
			numwant:    30.0,
			in:         sql.NullString{String: "xhtml/chapter3.xhtml", Valid: true},
			shouldFail: false,
		},
		{
			want:       "chapter01",
			numwant:    10.0,
			in:         sql.NullString{String: "ch01.html", Valid: true},
			shouldFail: false,
		},
		{
			want:       "chapter07",
			numwant:    70.0,
			in:         sql.NullString{String: "Text/chapter007.html", Valid: true},
			shouldFail: false,
		},
		{
			want:       "chapter18",
			numwant:    180.0,
			in:         sql.NullString{String: "xhtml/McKe_9780804137393_epub3_c018_r1.xhtm", Valid: true},
			shouldFail: false,
		},
		{
			want:       "preface03",
			numwant:    3.0,
			in:         sql.NullString{String: "preface03.html", Valid: true},
			shouldFail: false,
		},
		{
			want:       "introduction01",
			numwant:    1.0,
			in:         sql.NullString{String: "introduction1.html", Valid: true},
			shouldFail: false,
		},
	}

	for _, tt := range tests {
		sec, numsec := parseSection(tt.in)
		if (sec != tt.want || numsec != tt.numwant) && !tt.shouldFail {
			t.Errorf("Incorrect result: want: '%s', got: %s ; want: %0.2f, got: %0.2f", tt.want, sec, tt.numwant, numsec)
		}
	}
}
