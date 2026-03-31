package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func (app *App) printTextFile(filename string) error {
	content, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	if app.config.raw {
		_, err := os.Stdout.Write(content)
		return err
	}

	source := string(content)
	lexer := lexers.Match(filename)
	if lexer == nil {
		lexer = lexers.Analyse(source)
	}
	if lexer == nil {
		lexer = lexers.Fallback
	}

	iterator, err := chroma.Coalesce(lexer).Tokenise(nil, source)
	if err != nil {
		return err
	}

	style := styles.Get("catppuccin-frappe")
	if style == nil {
		style = styles.Fallback
	}
	lines := chroma.SplitTokensIntoLines(iterator.Tokens())
	lineDigits := 1
	if len(lines) > 0 {
		lineDigits = len(strconv.Itoa(len(lines)))
	}

	for i, line := range lines {
		fmt.Fprintf(os.Stdout, "%s %*d %s ", lineNumberPrefix(style), lineDigits, i+1, TERM_COLOR_RESET)
		if err := formatters.TTY16m.Format(os.Stdout, style, chroma.Literator(line...)); err != nil {
			return err
		}
		if !strings.HasSuffix(lastTokenValue(line), "\n") {
			fmt.Fprintln(os.Stdout)
		}
	}

	return nil
}

func lineNumberPrefix(style *chroma.Style) string {
	entry := style.Get(chroma.LineNumbers)
	var out strings.Builder

	if entry.Bold == chroma.Yes {
		out.WriteString("\033[1m")
	}
	if entry.Underline == chroma.Yes {
		out.WriteString("\033[4m")
	}
	if entry.Italic == chroma.Yes {
		out.WriteString("\033[3m")
	}
	if entry.Colour.IsSet() {
		fmt.Fprintf(&out, "\033[38;2;%d;%d;%dm", entry.Colour.Red(), entry.Colour.Green(), entry.Colour.Blue())
	}
	if entry.Background.IsSet() {
		fmt.Fprintf(&out, "\033[48;2;%d;%d;%dm", entry.Background.Red(), entry.Background.Green(), entry.Background.Blue())
	}

	return out.String()
}

func lastTokenValue(tokens []chroma.Token) string {
	if len(tokens) == 0 {
		return ""
	}
	return tokens[len(tokens)-1].Value
}
