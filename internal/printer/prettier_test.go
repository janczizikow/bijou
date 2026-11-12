package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettier(t *testing.T) {
	document := concat(
		text("bbbbb["),
		concat(
			concat(
				nest(2, concat(line(), text("ccc,"))),
				nest(2, concat(line(), text("dd"))),
			),
			nest(0, concat(line(), text("]"))),
		),
	)

	expected := "bbbbb[\n  ccc,\n  dd\n]"
	assert.Equal(t, expected, layout(document))
}
