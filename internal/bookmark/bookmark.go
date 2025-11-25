package bookmark

import (
	"database/sql"
	"fmt"
	"iter"
	"math"
	"regexp"
	"strconv"
	"strings"
)

const (
	MARKUP    = "markup"
	HIGHLIGHT = "highlight"
	NOTE      = "note"
	DOGEAR    = "dogear"
)

var (
	invalidCharsRgx = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]|(\.?x?html)`)
	// invalidCharsRgx = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)
	multiDashRgx = regexp.MustCompile(`-+`)
	locationRgx  = regexp.MustCompile(`(\\?\d+\\?\.\d+)`)
	chapterRgx   = regexp.MustCompile(`(chapter|ch|c)(\d+)|([a-z]+)(\d+)$`)
)

// Bookmarks are ordered by section and then location. However the location resets with every
// chapter, and sections coming before the book chapters (such as prefaces, introductions) have its
// own number. This means you can have
// - preface03, loc18.2
// - chapter01, loc 5.15
// Here chapter01 definitely comes after anything in the prefix, but if we compare the numbers alone
// the sorting will work backwards.
//
// This is a map that sets "multipliers" to each section, based on me knowing they'll always come
// e.g. at the beginning for "preface" or "introduction" and at the end for "apendix"
// The result should be a integer Identifier that should sort correctly:
// - preface03, loc18.2 => 3.182
// - chapter01, loc 5.15 => 10.515
// - appendix02, loc 20.9 => 40.209
//
// I doubt there are introductions/prefaces a number that reaches 10, but will see. Locations will
// be normalized to be less than one, so they are at the right of the decimal point
var sectionOrder = map[string]float64{
	"preface":      1.0,
	"prologue":     1.0,
	"introduction": 1.0,
	"chapter":      10.0,
	"appendix":     20.0,
}

type Bookmarks struct {
	Book       string
	Markups    []*Markup
	Highlights []*Highlight
}

func (self *Bookmarks) Marks() iter.Seq2[int, *Markup] {
	return func(yield func(idx int, mark *Markup) bool) {
		for i, m := range self.Markups {
			if !yield(i, m) {
				return
			}
		}
	}
}

func (self *Bookmarks) Highs() iter.Seq2[int, *Highlight] {
	return func(yield func(idx int, mark *Highlight) bool) {
		for i, h := range self.Highlights {
			if !yield(i, h) {
				return
			}
		}
	}
}

// Dummy interface to be able have fromRawValues returning either a Highlight or a Markup
type bookmark interface {
	Kind() string
}

func AllBookmarks(books []string) []*Bookmarks {
	all := make([]*Bookmarks, 0, 50)

	for _, b := range books {
		bms := &Bookmarks{
			Book:       b,
			Markups:    []*Markup{},
			Highlights: []*Highlight{},
		}

		raws := kdb.fetchBookmarks(b)

		for _, r := range raws {
			bm := fromRawValues(r)
			if bm != nil {
				switch bm.Kind() {
				case MARKUP:
					bms.Markups = append(bms.Markups, bm.(*Markup))
				case HIGHLIGHT:
					bms.Highlights = append(bms.Highlights, bm.(*Highlight))
				}
			}
		}

		all = append(all, bms)
	}

	return all
}

func AllBooks() []string {
	if books, err := kdb.fetchBooksWithBookmark(); err == nil {
		return books
	}
	return []string{}
}

// DOGEAR items are ignored: nil
func fromRawValues(kbm koboBookmark) bookmark {
	// If we can't understand the Type, we can't act on anything so we return early
	if !kbm.kind.Valid {
		return nil
	}
	orderId := 0.0
	// assuming a markup
	bm := &Markup{}
	if kbm.id.Valid {
		bm.Id = kbm.id.String
	}

	sec, numsec := parseSection(kbm.section)
	bm.Section = sec
	orderId = numsec

	loc, numloc := parseLocation(kbm.location)
	orderId += numloc
	bm.Location = loc
	// see "sectionOrder"
	bm.OrderId = orderId

	switch kbm.kind.String {
	case MARKUP:
		return bm

	case HIGHLIGHT, NOTE:
		if kbm.text.Valid && kbm.color.Valid {
			text := kbm.text.String
			col := kbm.color.Int64
			return &Highlight{
				Id:       bm.Id,
				Section:  bm.Section,
				Location: bm.Location,
				OrderId:  bm.OrderId,
				text:     text,
				color:    int(col),
			}
		}
	default:
		return nil
	}

	return nil
}

func sanitize(text string, noRepl bool) string {
	text = strings.ToLower(text)
	if text == "" || text == "none" {
		return "unknown"
	}
	repl := "-"
	if noRepl {
		repl = ""
	}
	valid := invalidCharsRgx.ReplaceAllString(text, repl)
	trimmed := strings.Trim(valid, "-")
	final := multiDashRgx.ReplaceAllString(trimmed, repl)

	return final
}

func parseSection(s sql.NullString) (sec string, numsec float64) {
	if !s.Valid {
		return "unknown00", 0.0
	}
	sect := sanitize(s.String, false)
	sect = strings.ToLower(sect)
	matches := chapterRgx.FindStringSubmatch(sect)

	if matches != nil {
		var numstr string
		var prefix string
		// section contains "chapter"
		if matches[1] != "" {
			prefix = "chapter"
			numstr = matches[2]
		} else if matches[3] != "" {
			// section contains something else e.g. "preface", "introduction"
			prefix = strings.ToLower(matches[3])
			numstr = matches[4]
		}
		num, err := strconv.Atoi(numstr)
		if err != nil {
			sec = sect
			num = 0
		}
		// see "sectionOrder"
		numsec += sectionOrder[prefix] * float64(num)
		sec = fmt.Sprintf("%s%02d", prefix, num)
	} else {
		sec = sect
	}
	return sec, numsec

}

func parseLocation(l sql.NullString) (string, float64) {
	if !l.Valid {
		return "loc0", 0.0
	}
	var numloc float64
	var loc string
	// formatting
	matches := locationRgx.FindStringSubmatch(l.String)
	if len(matches) > 1 {
		loc = sanitize(matches[1], true)
		num, _ := strconv.ParseFloat(loc, 64) // ignore error, keep whatever is default here
		numloc = num
	}
	// normalize so value is always < 1
	if numloc > 0 {
		intpart := math.Floor(numloc)
		numDigits := math.Floor(math.Log10(intpart)) + 1.0
		div := math.Pow(10, math.Max(numDigits, 3))
		numloc = numloc / div
		// round to 3 decimals
		ratio := math.Pow(10, float64(4))
		numloc = math.Round(numloc*ratio) / ratio
	}

	return loc, numloc
}
