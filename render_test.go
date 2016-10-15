package ml

import "testing"

func TestRenderMarkup(t *testing.T) {
	fooTpl := `
<foo>
	{{range .}}
		<bar />
	{{end}}
</foo>
	`

	data := []string{"bar1", "bar2"}

	res, err := renderMarkup(fooTpl, data)
	if err != nil {
		t.Error(err)
	}

	t.Log(res)

	if _, err = renderMarkup(fooTpl, 42); err == nil {
		t.Error("parse with number data should error")
	}

	invalidTpl := `
<foo>
	{{range .}}
		<bar />
	{{finish}}
</foo>
	`

	if _, err = renderMarkup(invalidTpl, data); err == nil {
		t.Error("parse with invalid template should error")
	}
}