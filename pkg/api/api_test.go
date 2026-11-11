package api_test

import (
	"testing"

	"github.com/janczizikow/bijou/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestValidateOptions(t *testing.T) {
	t.Parallel()
	opts := api.DefaultOptions()
	cases := []struct {
		name    string
		opts    api.FormatOptions
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name:    "valid options - no error",
			opts:    opts,
			wantErr: assert.NoError,
		},
		{
			name: "invalid tab width - errors",
			opts: api.FormatOptions{
				ArrowParens:   opts.ArrowParens,
				TrailingComma: opts.TrailingComma,
				ObjectWrap:    opts.ObjectWrap,
				TabWidth:      -1,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid print width - errors",
			opts: api.FormatOptions{
				ArrowParens:   opts.ArrowParens,
				TrailingComma: opts.TrailingComma,
				ObjectWrap:    opts.ObjectWrap,
				PrintWidth:    -1,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid arrow parens - errors",
			opts: api.FormatOptions{
				ArrowParens:   api.ArrowParens("test"),
				TrailingComma: opts.TrailingComma,
				ObjectWrap:    opts.ObjectWrap,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid trailing comma - errors",
			opts: api.FormatOptions{
				ArrowParens:   opts.ArrowParens,
				TrailingComma: api.TrailingComma("test"),
				ObjectWrap:    opts.ObjectWrap,
			},
			wantErr: assert.Error,
		},
		{
			name: "invalid object wrap - errors",
			opts: api.FormatOptions{
				ArrowParens:   opts.ArrowParens,
				TrailingComma: opts.TrailingComma,
				ObjectWrap:    api.ObjectWrap("test"),
			},
			wantErr: assert.Error,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tc.wantErr(t, tc.opts.Validate())
		})
	}
}
