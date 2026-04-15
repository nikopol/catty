package main

import (
	"flag"
	"fmt"
	"image/color"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
)

const version = "0.0.1"

type Config struct {
	debug bool
	raw   bool
	maxWidth int
	maxHeight int
	mimeType string
}

type App struct {
	config Config
}

func main() {
	app := App{
		config: Config{},
	}
	var versionMode bool
	flag.BoolVar(&app.config.debug, "d", false, "debug mode")
	flag.BoolVar(&app.config.raw, "r", false, "raw mode (no decoration)")
	flag.IntVar(&app.config.maxWidth, "w", 0, "max columns (default: terminal width)")
	flag.IntVar(&app.config.maxHeight, "h", 0, "max lines (used for image only, default: terminal height)")
	flag.StringVar(&app.config.mimeType, "m", "", "force file mime type")
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

	termWidth, termHeight, err := termSize()
	if err != nil {
		termWidth = 65535
		termHeight = 65535
		app.config.raw = true
	}
	if app.config.maxWidth == 0 {
		app.config.maxWidth = termWidth
	}
	if app.config.maxHeight == 0 {
		app.config.maxHeight = termHeight
	}

	app.printDebug("Output: %d x %d", app.config.maxWidth, app.config.maxHeight)

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

var debugColor = termFgColor(color.RGBA{0x88, 0x88, 0x88, 0xFF})
func (app *App) printDebug(txt string, args ...any) {
	if app.config.debug {
		fmt.Fprint(os.Stdout, debugColor)
		fmt.Fprintf(os.Stdout, txt, args...)
		fmt.Fprintln(os.Stdout, TERM_COLOR_RESET)
	}
}

func (app *App) printFile(filename string) error {
	mimeType := app.config.mimeType
	if mimeType == "" {
		mimeType = mimeTypeFromFilename(filename)
	}
	app.printDebug("File: %s\nMime Type: %s", filename, mimeType)
	if strings.HasPrefix(mimeType, "image/") || mimeType == "image" {
		app.printDebug("image detected")
		return app.printImageFile(filename)
	}
	if strings.HasPrefix(mimeType, "text/") || mimeType == "text" || isTextFile(filename, mimeType) {
		app.printDebug("text detected")
		return app.printTextFile(filename)
	}
	app.printDebug("binary detected")
	return app.printBinaryFile(filename)
}

func mimeTypeFromFilename(filename string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}
