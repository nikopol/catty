package main

import (
	"errors"
	"fmt"
	"image/color"
	"io"
	"os"
	"strconv"
	"strings"
)

func (app *App) printBinaryFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	prefixWidth := 12 // " 0000000000 "
	bytesPerLine := (app.config.width - prefixWidth - 6) / 4
	if bytesPerLine < 1 {
		return errors.New("term too small")
	}

	buf := make([]byte, 2048)
	line := make([]byte, 0, bytesPerLine)
	lineOffset := 0

	termIndexPanelColor := termFgBgColor(
		color.RGBA{0xA0,0xA0,0xA0,0xFF},
		color.RGBA{0x20,0x20,0x66,0xFF},
	)
	termHexDumpPanelColor := termFgBgColor(
		color.RGBA{0xFF,0xFF,0xFF,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)
	termAsciiDumpPanelColor := termFgBgColor(
		color.RGBA{0xC0,0xC0,0xC0,0xFF},
		color.RGBA{0x20,0x20,0x66,0xFF},
	)
	termAsciiUnprintableColor := termFgColor(
		color.RGBA{0xFF,0x66,0x66,0xFF},
	)

	flushLine := func() {
		if len(line) == 0 {
			return
		}
		fillGap := bytesPerLine - len(line)

		var out strings.Builder
		// INDEX PANEL
		if app.config.raw {
			fmt.Fprintf(&out, " %10x ", lineOffset)
		} else {
			fmt.Fprintf(&out, "%s %10d %s", termIndexPanelColor, lineOffset, TERM_COLOR_RESET)
		}
		// HEX DUMP PANEL
		if !app.config.raw {
			fmt.Fprint(&out, termHexDumpPanelColor)
		}
		fmt.Fprint(&out, "│")
		for _, b := range line {
			fmt.Fprintf(&out, " %02x", b)
		}
		for i := 0; i < fillGap; i++ {
			out.WriteString("   ")
		}
		out.WriteString(" │")
		// ASCII DUMP PANEL
		if !app.config.raw {
			fmt.Fprint(&out, termAsciiDumpPanelColor)
		}
		out.WriteString(" ")
		for _, b := range line {
			if strconv.IsPrint(rune(b)) {
				out.WriteByte(b)
			} else {
				if !app.config.raw {
					fmt.Fprint(&out, termAsciiUnprintableColor)
					out.WriteByte('.')
					fmt.Fprint(&out, termAsciiDumpPanelColor)
				} else {
					out.WriteByte('.')
				}
			}
		}
		for i := 0; i < fillGap; i++ {
			out.WriteByte(' ')
		}
		out.WriteByte(' ')
		if !app.config.raw {
			fmt.Fprint(&out, TERM_COLOR_RESET)
		}
		fmt.Println(out.String())
		lineOffset += len(line)
		line = line[:0]
	}

	for {
		n, err := file.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				line = append(line, b)
				if len(line) == bytesPerLine {
					flushLine()
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	flushLine()

	return nil
}
