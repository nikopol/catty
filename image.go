package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
)

func (app *App) printImageFile(filename string) error {
	img, err := readImageFile(filename)
	if err != nil {
		return err
	}
	app.printImage(img)
	return nil
}

func readImageFile(filename string) (image.Image, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func medianColor(colors []color.Color) color.Color {
	count := len(colors)
	if count == 0 {
		return color.RGBA{}
	}

	reds := 0
	greens := 0
	blues := 0
	alphas := 0

	for _, c := range colors {
		r, g, b, a := c.RGBA()
		reds += int(r >> 8)
		greens += int(g >> 8)
		blues += int(b >> 8)
		alphas += int(a >> 8)
	}

	return color.RGBA{
		R: uint8(reds / count),
		G: uint8(greens / count),
		B: uint8(blues / count),
		A: uint8(alphas / count),
	}
}

func resizeToMaxWidth(img image.Image, maxWidth int, addYScale float64) *image.RGBA {
	bounds := img.Bounds()
	srcWidth := bounds.Dx()
	srcHeight := bounds.Dy()
	width := min(maxWidth, srcWidth)

	if width <= 0 || srcWidth == 0 || srcHeight == 0 {
		return image.NewRGBA(image.Rect(0, 0, 0, 0))
	}

	scale := float64(width) / float64(srcWidth)
	height := max(1, int(float64(srcHeight)*scale*addYScale))

	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xScale := float64(srcWidth) / float64(width)
	yScale := float64(srcHeight) / float64(height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var rgba color.Color
			if scale > 0.5 {
				srcX := bounds.Min.X + int(float64(x)*xScale)
				srcY := bounds.Min.Y + int(float64(y)*yScale)
				rgba = img.At(srcX, srcY)
			} else {
				srcXStart := bounds.Min.X + int(float64(x)*xScale)
				srcXEnd := bounds.Min.X + int(float64(x+1)*xScale)
				srcYStart := bounds.Min.Y + int(float64(y)*yScale)
				srcYEnd := bounds.Min.Y + int(float64(y+1)*yScale)
				if srcXEnd <= srcXStart {
					srcXEnd = srcXStart + 1
				}
				if srcYEnd <= srcYStart {
					srcYEnd = srcYStart + 1
				}
				values := make([]color.Color, 0, (srcXEnd-srcXStart)*(srcYEnd-srcYStart))
				for srcY := srcYStart; srcY < srcYEnd; srcY++ {
					for srcX := srcXStart; srcX < srcXEnd; srcX++ {
						values = append(values, img.At(srcX, srcY))
					}
				}
				rgba = medianColor(values)
			}
			dst.Set(x, y, rgba)
		}
	}

	return dst
}

func (app *App) printImage(img image.Image) {
	maxWidth := app.config.width
	termYScale := 5.0 / 9.0
	resized := resizeToMaxWidth(img, maxWidth, termYScale)
	bounds := resized.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		var line strings.Builder
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba := resized.At(x, y)
			_, _, _, a := rgba.RGBA()
			if a == 0 {
				line.WriteRune(' ')
				continue
			}
			line.WriteString(termFgColor(rgba))
			line.WriteRune('█')
		}
		line.WriteString(TERM_COLOR_RESET)
		fmt.Println(line.String())
	}
}
