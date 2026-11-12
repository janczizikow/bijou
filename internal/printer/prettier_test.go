package printer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type Node struct {
	Val      string
	Children []Node
}

func showBracket(ts []Node) Doc {
	if len(ts) == 0 {
		return nil
	}
	return concat(
		text("["),
		concat(nest(2, concat(line(), showTrees(ts))), concat(line(), text("]"))),
	)
}

func showTrees(ts []Node) Doc {
	if len(ts) == 0 {
		return nil
	}
	if len(ts) == 1 {
		return showTree(ts[0])
	}
	return concat(showTree(ts[0]), concat(text(","), concat(line(), showTrees(ts[1:]))))
}

func showTree(t Node) Doc {
	return concat(text(t.Val), showBracket(t.Children))
}

func TestPrettier(t *testing.T) {

	tree := Node{
		Val:      "bbbbb",
		Children: []Node{{Val: "ccc"}, {Val: "dd"}},
	}

	document := showTree(tree)
	// concat(
	// 	text("bbbbb["),
	// 	concat(
	// 		concat(
	// 			nest(2, concat(line(), text("ccc,"))),
	// 			nest(2, concat(line(), text("dd"))),
	// 		),
	// 		nest(0, concat(line(), text("]"))),
	// 	),
	// )

	expected := "bbbbb[\n  ccc,\n  dd\n]"
	assert.Equal(t, expected, layout(document))
}
