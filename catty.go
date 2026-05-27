package main

import (
	"bytes"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"image/color"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
)

const version = "1.01"

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
	var forceImageMode bool
	var forceBinaryMode bool
	flag.BoolVar(&app.config.debug, "d", false, "debug mode")
	flag.BoolVar(&app.config.raw, "r", false, "raw mode (no decoration)")
	flag.BoolVar(&app.config.raw, "p", false, "alias for -r (no decoration)")
	flag.IntVar(&app.config.maxWidth, "w", 0, "max columns (default: terminal width)")
	flag.IntVar(&app.config.maxHeight, "h", 0, "max lines (used for image only, default: terminal height)")
	flag.StringVar(&app.config.mimeType, "m", "", "force file mime type")
	flag.BoolVar(&versionMode, "v", false, "show version")
	flag.BoolVar(&forceBinaryMode, "bin", false, "force hex dump mode (alias for -m bin)")
	flag.BoolVar(&forceImageMode, "img", false, "force image mode (alias for -m image)")
	flag.Parse()

	if versionMode {
		fmt.Printf("catty v%s\n", version)
		os.Exit(0)
	}
	if forceBinaryMode {
		app.config.mimeType = "binary"
	}
	if forceImageMode {
		app.config.mimeType = "image"
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

	if hasStdinPipe() {
		if err := app.printByData(os.Stdin); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(9)
		}
		os.Exit(0)
	} else if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "what file?")
		os.Exit(1)
	}

	for i := 0; i < flag.NArg(); i++ {
		if err := app.printByFilename(flag.Arg(i)); err != nil {
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

func hasStdinPipe() bool {
	return !isatty.IsTerminal(os.Stdin.Fd()) && !isatty.IsCygwinTerminal(os.Stdin.Fd())
}

func (app *App) printByFilename(filename string) error {
	mimeType := app.config.mimeType
	if mimeType == "" {
		mimeType = mimeTypeFromFilename(filename)
	}
	app.printDebug("File: %s\nMime Type: %s", filename, mimeType)
	if strings.HasPrefix(mimeType, "image") {
		app.printDebug("image detected")
		return app.printImageFile(filename)
	}
	if strings.HasPrefix(mimeType, "text") || isTextFile(filename, mimeType) {
		app.printDebug("text detected")
		return app.printTextFile(filename)
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	return app.printByData(f)
}

func (app *App) printByData(file *os.File) error {
	data, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	mimeType := app.config.mimeType
	if mimeType == "" {
		mimeType = http.DetectContentType(data)
	}
	app.printDebug("Mime Type: %s", mimeType)

	if strings.HasPrefix(mimeType, "image") {
		app.printDebug("image detected")
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return err
		}
		app.printImage(img)
		return nil
	}
	if strings.HasPrefix(mimeType, "text") {
		app.printDebug("text detected")
		return app.printTextContent(data, "")
	}
	app.printDebug("binary detected")
	return app.printBinaryReader(bytes.NewReader(data))
}

func mimeTypeFromFilename(filename string) string {
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		return "application/octet-stream"
	}
	return mimeType
}
