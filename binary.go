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
	return app.printBinaryReader(file)
}

func (app *App) printBinaryReader(r io.Reader) error {
	if app.config.raw {
		return printRawBinaryFile(r)
	}

	prefixWidth := 12 // " 0000000000 "
	bytesPerLine := (app.config.maxWidth - prefixWidth - 6) / 4
	if bytesPerLine < 1 {
		return errors.New("term too small")
	}

	buf := make([]byte, 2048)
	line := make([]byte, 0, bytesPerLine)
	lineOffset := 0

	termIndexPanelColor := termFgBgColor(
		color.RGBA{0x70,0x80,0x9C,0xFF},
		color.RGBA{0x36,0x43,0x59,0xFF},
	)
	termHexDumpPanelColor := termFgBgColor(
		color.RGBA{0xFF,0xFF,0xFF,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)
	termASCIIColor0 := termFgBgColor(
		color.RGBA{0x88,0x88,0x88,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)
	termASCIIColor1_19 := termFgBgColor(
		color.RGBA{0x47,0xB7,0x62,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)
	termASCIIColor20_7E := termFgBgColor(
		color.RGBA{0x6E,0xAD,0xBC,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)
	termASCIIColor7F_FE := termFgBgColor(
		color.RGBA{0xC1,0xC7,0x75,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)
	termASCIIColorFF := termFgBgColor(
		color.RGBA{0xFF,0xFF,0xFF,0xFF},
		color.RGBA{0x00,0x00,0x00,0xFF},
	)

	getByteColor := func(b byte) string {
		if b == 0 {
			return termASCIIColor0;
		} else if b < 0x20 {
			return termASCIIColor1_19;
		} else if b < 0x7F {
			return termASCIIColor20_7E;
		} else if b < 0xFF {
			return termASCIIColor7F_FE;
		} else {
			return termASCIIColorFF;
		}
	}

	flushLine := func() {
		if len(line) == 0 {
			return
		}
		fillGap := bytesPerLine - len(line)

		var out strings.Builder
		// INDEX PANEL
		fmt.Fprintf(&out, "%s %10d %s", termIndexPanelColor, lineOffset, TERM_COLOR_RESET)
		// HEX DUMP PANEL
		fmt.Fprint(&out, termHexDumpPanelColor)
		fmt.Fprint(&out, "│")
		for _, b := range line {
			fmt.Fprintf(&out, "%s %02X", getByteColor(b), b)
		}
		for i := 0; i < fillGap; i++ {
			out.WriteString("   ")
		}
		fmt.Fprintf(&out, " %s│ ", termHexDumpPanelColor)
		// ASCII DUMP PANEL
		for _, b := range line {
			fmt.Fprint(&out, getByteColor(b))
			if strconv.IsPrint(rune(b)) {
				out.WriteByte(b)
			} else {
				out.WriteString("•")
			}
		}
		for i := 0; i < fillGap; i++ {
			out.WriteByte(' ')
		}
		out.WriteByte(' ')
		fmt.Fprint(&out, TERM_COLOR_RESET)
		fmt.Println(out.String())
		lineOffset += len(line)
		line = line[:0]
	}

	for {
		n, err := r.Read(buf)
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

func printRawBinaryFile(r io.Reader) error {
	buf := make([]byte, 2048)
	first := true

	for {
		n, err := r.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				if !first {
					fmt.Fprint(os.Stdout, " ")
				}
				fmt.Fprintf(os.Stdout, "%02x", b)
				first = false
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}

	if !first {
		fmt.Fprintln(os.Stdout)
	}

	return nil
}
