package bookmark

import (
	"fmt"
	"os"
	"path/filepath"
)

type Markup struct {
	Id        string
	BookTitle string
	Section   string
	Part      string
	Location  string
	OrderId   float64
	svgPath   string
	jpgPath   string
}

func (self *Markup) Kind() string {
	return MARKUP
}

func (self *Markup) Outfile() string {
	return fmt.Sprintf(
		"mk_%s_%s_%s_%f.png",
		self.Id[:8],
		self.Section,
		self.Location,
		self.OrderId,
	)
}

// Compute SVG file path base on the Bookmark ID
func (self *Markup) SvgFile(markPath string) string {
	if self.svgPath != "" {
		return self.svgPath
	}
	self.svgPath = filepath.Join(markPath, fmt.Sprintf("%s.svg", self.Id))
	return self.svgPath
}

// Compute JPG file path base on the Bookmark ID
func (self *Markup) JpgFile(markPath string) string {
	if self.jpgPath != "" {
		return self.jpgPath
	}
	self.jpgPath = filepath.Join(markPath, fmt.Sprintf("%s.jpg", self.Id))
	return self.jpgPath
}

func (self *Markup) HasImagePair(markPath string) bool {
	jf := self.JpgFile(markPath)
	sf := self.SvgFile(markPath)
	if jf == "" || sf == "" {
		return false
	}
	if _, err := os.Stat(jf); os.IsNotExist(err) {
		return false
	}
	if _, err := os.Stat(sf); os.IsNotExist(err) {
		return false
	}

	return true
}
