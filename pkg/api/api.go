package api

import (
	"fmt"
	"strings"

	"github.com/janczizikow/bijou/internal/parser"
	"github.com/janczizikow/bijou/internal/printer"
)

type ArrowParens string

const (
	ArrowParensAlways ArrowParens = "always"
	ArrowParensAvoid  ArrowParens = "avoid"
)

func ParseArrowParens(value string) (ArrowParens, error) {
	switch strings.ToLower(value) {
	case "always", "":
		return ArrowParensAlways, nil
	case "avoid":
		return ArrowParensAvoid, nil
	default:
		return "", fmt.Errorf("invalid arrow-parens %q (allowed: %s, %s)", value, ArrowParensAlways, ArrowParensAvoid)
	}
}

type TrailingComma string

const (
	TrailingCommaAll  = "all"
	TrailingCommaES5  = "es5"
	TrailingCommaNone = "none"
)

func ParseTrailingComma(value string) (TrailingComma, error) {
	switch strings.ToLower(value) {
	case "es5", "":
		return TrailingCommaES5, nil
	case "all":
		return TrailingCommaAll, nil
	case "node":
		return TrailingCommaNone, nil
	default:
		return "", fmt.Errorf("invalid trailing comma %q (allowed: %s, %s, %s)",
			value, TrailingCommaAll, TrailingCommaES5, TrailingCommaNone)
	}
}

type ObjectWrap string

const (
	ObjectWrapPreserve = "preserve"
	ObjectWrapCollapse = "collapse"
)

func ParseObjectWrap(value string) (ObjectWrap, error) {
	switch strings.ToLower(value) {
	case "preserve", "":
		return ObjectWrapPreserve, nil
	case "collapse":
		return ObjectWrapCollapse, nil
	default:
		return "", fmt.Errorf("invalid object wrap %q (allowed: %s, %s)",
			value, ObjectWrapPreserve, ObjectWrapCollapse)
	}
}

type FormatOptions struct {
	ArrowParens        ArrowParens
	BracketSameLine    bool
	UseTabs            bool
	TabWidth           int
	PrintWidth         int
	SingleQuote        bool
	TrailingComma      TrailingComma
	BracketSpacing     bool
	JsxBracketSameLine bool
	Semi               bool
	ObjectWrap         ObjectWrap
}

func (o FormatOptions) Validate() error {
	if o.TabWidth < 0 {
		return fmt.Errorf("tab width cannot be negative: %d", o.TabWidth)
	}

	if o.PrintWidth < 0 {
		return fmt.Errorf("print width cannot be negative: %d", o.PrintWidth)
	}

	switch o.ArrowParens {
	case ArrowParensAlways, ArrowParensAvoid:
	default:
		return fmt.Errorf("invalid arrow-parens value %q", o.ArrowParens)
	}

	switch o.TrailingComma {
	case TrailingCommaAll, TrailingCommaES5, TrailingCommaNone:
	default:
		return fmt.Errorf("invalid trailing comma value %q", o.TrailingComma)
	}

	switch o.ObjectWrap {
	case ObjectWrapPreserve, ObjectWrapCollapse:
	default:
		return fmt.Errorf("invalid object wrap value %q", o.ObjectWrap)
	}

	return nil
}

func toPrinterOptions(options FormatOptions) printer.Options {
	return printer.Options{
		LineLimit:   options.PrintWidth,
		Semi:        options.Semi,
		SingleQuote: options.SingleQuote,
		UseTabs:     options.UseTabs,
		TabWidth:    options.TabWidth,
		// TrailingComma
		BracketSpacing: options.BracketSpacing,
	}
}

func DefaultOptions() FormatOptions {
	return FormatOptions{
		ArrowParens:        ArrowParensAlways,
		BracketSameLine:    false,
		UseTabs:            false,
		TabWidth:           2,
		PrintWidth:         80,
		SingleQuote:        false,
		TrailingComma:      TrailingCommaES5,
		BracketSpacing:     true,
		JsxBracketSameLine: false,
		Semi:               true,
		ObjectWrap:         ObjectWrapPreserve,
	}
}

func Format(input string, options FormatOptions) (string, error) {
	if err := options.Validate(); err != nil {
		return "", err
	}
	ast, err := parser.Parse(input)
	if err != nil {
		return "", err
	}

	res := printer.Print(ast, toPrinterOptions(options))

	return string(res), nil
}

func FormatJSON(input string, options FormatOptions) (string, error) {
	if err := options.Validate(); err != nil {
		return "", err
	}
	expr, err := parser.ParseJSON(input)
	if err != nil {
		return "", err
	}
	opts := toPrinterOptions(options)
	opts.SingleQuote = false
	res := printer.PrintExpression(expr, opts)

	return string(res), nil
}
