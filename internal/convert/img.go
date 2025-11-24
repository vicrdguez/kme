package convert

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"kme/internal/bookmark"
	"os"
	"path/filepath"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

const (
	WIDTH  = 1264
	HEIGHT = 1680
)

// Global to store font
var defFont *sfnt.Font

func svgToRGBA(svgPath string) (image.Image, error) {
	file, err := os.Open(svgPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SVG file %s: %w", svgPath, err)
	}
	defer file.Close()

	icon, err := oksvg.ReadIconStream(file, oksvg.WarnErrorMode)
	if err != nil {
		return nil, fmt.Errorf("failed to read SVG stream %s: %w", svgPath, err)
	}

	// Set viewBox to match target dimensions if not present, or scale to fit
	icon.SetTarget(0, 0, float64(WIDTH), float64(HEIGHT))

	// Create new transparent base img
	img := image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT))
	// rasterizer
	scanner := rasterx.NewScannerGV(int(icon.ViewBox.W), int(icon.ViewBox.H), img, img.Bounds())
	raster := rasterx.NewDasher(WIDTH, HEIGHT, scanner)
	// draw svg strokes in base img
	icon.Draw(raster, 1.0) // opacity 1.0

	return img, nil
}

func OverlayMarkup(m *bookmark.Markup, markPath string, outPath string) error {
	if !m.HasImagePair(markPath) {
		return fmt.Errorf("\tThis bookmark does not have both needed Markup files: %s", m.Id)
	}
	markImg, err := svgToRGBA(m.SvgFile(markPath))
	if err != nil {
		return fmt.Errorf("Failed to render SVG: %v", err)
	}

	base, err := os.Open(m.JpgFile(markPath))
	if err != nil {
		return fmt.Errorf("Failed to open background image %s: %w", m.JpgFile(markPath), err)
	}
	defer base.Close()

	baseImg, err := jpeg.Decode(base)
	if err != nil {
		return fmt.Errorf("Failed to decode base image %s: %w", m.JpgFile(markPath), err)
	}

	// new RGBA img for the background, with the intended size: The canvas
	container := image.NewRGBA(image.Rect(0, 0, WIDTH, HEIGHT))

	// b := baseImg.Bounds().Size()
	draw.Draw(container, container.Bounds(), baseImg, image.Point{}, draw.Src)
	// Draw the original decoded base img over the background canvas
	// This will fit the base img into the WxH
	// Maybe use ApproxBiLinear for better performance if quality does not suffer much
	// draw.ApproxBiLinear.Scale(container, container.Bounds(), baseImg, baseImg.Bounds(), draw.Over, nil)

	// Add text to the bg img
	text := fmt.Sprintf("%s/%s", m.Section, m.Location)
	textColor := color.Black
	// Pos: 10 (from left), 25 (from top)
	// This is the base location to write the text from
	point := fixed.Point26_6{X: fixed.Int26_6(10 * 64), Y: fixed.Int26_6(25 * 64)}

	df := &font.Drawer{
		Dst:  container,
		Src:  image.NewUniform(textColor),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	// Draw the text
	df.DrawString(text)

	// overlay the markup img with the background img
	draw.Draw(container, markImg.Bounds(), markImg, image.Point{}, draw.Over)

	imgPath := filepath.Join(outPath, m.Outfile())

	return encode(imgPath, container)
}

func encode(outPath string, img *image.RGBA) error {
	out, err := os.Create(outPath)
	if err != nil {
		return fmt.Errorf("Failed to create output file %s: %w", outPath, err)
	}
	defer out.Close()

	// encode final image into the output
	if err := jpeg.Encode(out, img, &jpeg.Options{Quality: 90}); err != nil {
		return fmt.Errorf("jpeg encode failed: %w", err)
	}

	// if BestCompression is too slow, could try with Default or BestSpeed
	// encoder := png.Encoder{CompressionLevel: png.BestCompression}
	// if err := encoder.Encode(out, bgImg); err != nil {
	// 	return fmt.Errorf("Failed to encode final PNG output to %s: %w", outPath, err)
	// }

	fmt.Println("\t Saving final overlay: ", out.Name())

	return nil
}
