package api

import (
	"github.com/janczizikow/bijou/internal/parser"
	"github.com/janczizikow/bijou/internal/printer"
)

type FormatOptions struct {
	UseTabs            bool
	TabWidth           int
	PrintWidth         int
	SingleQuote        bool
	TrailingComma      string
	BracketSpacing     bool
	JsxBracketSameLine bool
	Semi               bool
}

func Format(input string, options FormatOptions) (string, error) {
	ast, err := parser.Parse(input)
	if err != nil {
		return "", err
	}

	res := printer.Print(ast, toPrinterOptions(options))

	return string(res), nil
}

func toPrinterOptions(options FormatOptions) printer.Options {
	return printer.Options{
		LineLimit:   options.PrintWidth,
		Semi:        options.Semi,
		SingleQuote: options.SingleQuote,
	}
}
