package bookmark

import (
	"fmt"
	"strings"
)

const (
	green = '🟩'
	blue = '🟦'
	red = '🟥'
	yellow = '🟨'
)

type Highlight struct {
	Id        string
	BookTitle string
	Section   string
	Part      string
	Location  string
	OrderId   float64
	text      string
	color     int
}

func (self *Highlight) Kind() string {
	return HIGHLIGHT
}

// The Kobo DB only provides with the color code as int, and the name does not seem available in
// any other table. So I'm hardoding the colors here, since it is small enough and quicker than
// going for it to the DB anyways if it was even possible.
// Found this mapping here: https://www.mobileread.com/forums/showthread.php?t=366073
// Just trusting this bro is right
func (self *Highlight) Colors(code int) rune {
	// Original color code mapping
	// 0: "yellow"
	// 1: "red"
	// 2: "blue"
	// 3: "green"

	// This are the colors I want as CSS outputs for the background and font
	// [0] bg color, [1] font color
	colors := map[int]rune{
		0: yellow,
		1: red,
		2: blue,
		3: green,
	}
	return colors[code]
}

func (self *Highlight) Format() string {
	format := `%c %s --- %s`
	loc := fmt.Sprintf("%s.%s", self.Section, self.Location)
	color := self.Colors(self.color)
	return fmt.Sprintf(format, color, strings.TrimSpace(self.text), loc)
}
