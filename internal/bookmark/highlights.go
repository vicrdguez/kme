package bookmark

import (
	"fmt"
	"strings"
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
func (self *Highlight) Colors(code int) []string {
	// Original color code mapping
	// 0: "yellow"
	// 1: "red"
	// 2: "blue"
	// 3: "green"

	// This are the colors I want as CSS outputs for the background and font
	// [0] bg color, [1] font color
	colors := map[int][]string{
		0: {"gold", "black"},
		1: {"crimson", "white"},
		2: {"darkcyan", "white"},
		3: {"lightgreen", "black"},
	}
	return colors[code]
}

func (self *Highlight) Format() string {
	format := `<span style='background-color: %s; color: %s'>%s</span> --- %s`
	loc := fmt.Sprintf("%s.%s", self.Section, self.Location)
	colors := self.Colors(self.color)
	return fmt.Sprintf(format, colors[0], colors[1], strings.TrimSpace(self.text), loc)
}
