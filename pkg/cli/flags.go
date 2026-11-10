package cli

import (
	"github.com/janczizikow/bijou/pkg/api"
)

type arrowParensValue struct {
	target *api.ArrowParens
}

func (ap *arrowParensValue) String() string {
	if ap.target == nil {
		return ""
	}
	return string(*ap.target)
}

func (ap *arrowParensValue) Set(value string) error {
	v, err := api.ParseArrowParens(value)
	if err != nil {
		return err
	}
	ap.target = &v
	return nil
}

type trailingCommaValue struct {
	target *api.TrailingComma
}

func (tc *trailingCommaValue) String() string {
	if tc.target == nil {
		return ""
	}
	return string(*tc.target)
}

func (tc *trailingCommaValue) Set(value string) error {
	v, err := api.ParseTrailingComma(value)
	if err != nil {
		return err
	}
	tc.target = &v
	return nil
}
