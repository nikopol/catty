package main

import (
	"flag"
	"fmt"
	"image/color"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/mattn/go-isatty"
)

const version = "0.0.1"

type Config struct {
	debug bool
	raw   bool
	width int
}

type App struct {
	config Config
}

var debugColor = termFgColor(color.RGBA{0x88, 0x88, 0x88, 0xFF})

func main() {
	app := App{
		config: Config{},
	}
	var versionMode bool
	flag.BoolVar(&app.config.debug, "d", false, "debug mode")
	flag.BoolVar(&app.config.raw, "r", false, "raw mode (no decoration)")
	flag.IntVar(&app.config.width, "w", 0, "max columns (default: terminal width)")
	flag.BoolVar(&versionMode, "v", false, "show version")
	flag.Parse()

	if versionMode {
		fmt.Printf("catty v%s\n", version)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "what file?")
		os.Exit(1)
	}

	if app.config.width == 0 {
		width, _, err := termSize()
		if err != nil {
			width = 65535
			app.config.raw = true
		}
		app.config.width = width
		app.printDebug("Output width: %d", width)
	}

	if !app.config.raw && !isatty.IsTerminal(os.Stdout.Fd()) {
		app.config.raw = true
	}

	for i := 0; i < flag.NArg(); i++ {
		if err := app.printFile(flag.Arg(i)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(9)
		}
	}
	os.Exit(0)
}

func (app *App) printDebug(txt string, args ...any) {
	if app.config.debug {
		fmt.Fprint(os.Stdout, debugColor)
		fmt.Fprintf(os.Stdout, txt, args...)
		fmt.Fprintln(os.Stdout, TERM_COLOR_RESET)
	}
}

func (app *App) printFile(filename string) error {
	mimeType := mimeTypeFromFilename(filename)
	app.printDebug("File: %s\nMime Type: %s", filename, mimeType)
	if strings.HasPrefix(mimeType, "image/") {
		return app.printImageFile(filename)
	}
	if strings.HasPrefix(mimeType, "text/") || isTextFile(filename, mimeType) {
		return app.printTextFile(filename)
	}
	return app.printBinaryFile(filename)
}

func mimeTypeFromFilename(filename string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}

func isTextFile(filename string, mimeType string) bool {
	switch mimeType {
	case "application/json", "application/xml", "application/javascript", "application/typescript", "application/x-typescript":
		return true
	}

	return lexers.Match(filename) != nil
}
