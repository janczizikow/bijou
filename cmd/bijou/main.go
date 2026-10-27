package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/janczizikow/bijou/pkg/api"
	"github.com/janczizikow/bijou/pkg/cli"
)

const version = "1.0.0"

func main() {
	printVersion := flag.Bool("version", false, "print bijou version")

	tabs := flag.Bool("use-tabs", false, "Indent with tabs instead of spaces\n")
	tabWidth := flag.Int("tab-width", 2, "Number of spaces per indentation level\n")
	printWidth := flag.Int("print-width", 80, "The line length to wrap\n")
	singleQuote := flag.Bool("single-quote", false, "Use single quotes instead of double quotes\n")
	trailingComma := flag.String("trailing-comma", "es5", "Print trailing commas wherever possible when multi-line\n")
	bracketSpacing := flag.Bool("bracket-spacing", true, "Print spaces between brackets\n")
	jsxBracketSameLine := flag.Bool("jsx-bracket-same-line", false, "Print > on the last line instead of at a new line\n")
	semi := flag.Bool("semi", true, "Print semicolons\n")
	// TODO: verify some of the bool options (some of them seem to be the opposite in prettier). e.g.: --no-semi)
	flag.Parse()

	opts := api.FormatOptions{
		UseTabs:            *tabs,
		TabWidth:           *tabWidth,
		PrintWidth:         *printWidth,
		SingleQuote:        *singleQuote,
		TrailingComma:      *trailingComma,
		BracketSpacing:     *bracketSpacing,
		JsxBracketSameLine: *jsxBracketSameLine,
		Semi:               *semi,
	}
	app := cli.New(opts)

	if printVersion != nil && *printVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	if flag.NArg() == 0 {
		fmt.Println("Usage: bijou [options] [file/dir/glob ...]")
		fmt.Printf("\noptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	os.Exit(app.Run(flag.Args()))
}
