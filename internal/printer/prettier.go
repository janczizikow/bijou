package printer

import "strings"

type Doc interface {
	doc()
}

func (txt *Text) doc() {}
func (ln *Line) doc()  {}
func (u *Union) doc()  {}

type Text struct {
	value string
	next  Doc
}

type Line struct {
	indent int
	next   Doc
}

type Union struct {
	left  Doc
	right Doc
}

func text(s string) Doc {
	if s == "" {
		return nil
	}
	return &Text{value: s, next: nil}
}

func line() Doc {
	return &Line{}
}

func concat(l, r Doc) Doc {
	if l == nil {
		return r
	}
	return &Union{left: l, right: r}
}

func nest(i int, doc Doc) Doc {
	if i == 0 {
		return doc
	}
	switch v := doc.(type) {
	case *Union:
		// nest i (x <> y) = nest i x <> nest i y
		return &Union{left: nest(i, v.left), right: nest(i, v.right)}
	case *Text:
		// nest i (text s) = text s
		return v
	case *Line:
		// nest (i+j) x = nest i (nest j x)
		return &Line{indent: v.indent + i, next: v.next}
	case nil:
		// nest i nil = nil
		return nil
	}
	return doc
}

func layout(doc Doc) string {
	switch v := doc.(type) {
	case *Union:
		return layout(v.left) + layout(v.right)
	case *Text:
		return v.value + layout(v.next)
	case *Line:
		return "\n" + strings.Repeat(" ", v.indent)
	case nil:
		return ""
	}

	return ""
}
