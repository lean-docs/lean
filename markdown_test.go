package lean_test

import (
	"bytes"
	"os"
	"strings"
	"testing"

	lean "github.com/lean-docs/lean"
)

func TestMarkdownSourcePreservesBytes(t *testing.T) {
	fixture, err := os.ReadFile("testdata/fixtures/markdown/table.md")
	if err != nil {
		t.Fatal(err)
	}
	source := []byte("\ufeff" + strings.ReplaceAll(string(fixture), "\n", "\r\n"))
	original := bytes.Clone(source)
	document, err := lean.OpenMarkdownSource(source)
	if err != nil {
		t.Fatal(err)
	}
	source[0] = 0
	if !bytes.Equal(document.Bytes(), original) {
		t.Fatal("opening source did not isolate the original bytes")
	}
	copy := document.Bytes()
	copy[0] = 0
	if !bytes.Equal(document.Bytes(), original) {
		t.Fatal("reading source changed the original bytes")
	}
	if _, err := document.RenderHTML(); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(document.Bytes(), original) {
		t.Fatal("rendering changed source bytes")
	}
}

func TestMarkdownSourceRendersUpstreamFixtures(t *testing.T) {
	for _, name := range []string{"table", "tasklist", "strikethrough", "headings", "code"} {
		t.Run(name, func(t *testing.T) {
			source, err := os.ReadFile("testdata/fixtures/markdown/" + name + ".md")
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile("testdata/fixtures/markdown/" + name + ".html")
			if err != nil {
				t.Fatal(err)
			}
			document, err := lean.OpenMarkdownSource(source)
			if err != nil {
				t.Fatal(err)
			}
			actual, err := document.RenderHTML()
			if err != nil || !bytes.Equal(actual, expected) {
				t.Fatalf("unexpected preview: %s (%v)", actual, err)
			}
		})
	}
}

func TestMarkdownSourceRejectsInvalidText(t *testing.T) {
	for _, input := range [][]byte{{0xff}} {
		if _, err := lean.OpenMarkdownSource(input); err != lean.ErrInvalidMarkdown {
			t.Fatalf("invalid text was accepted: %v", err)
		}
	}
	for _, input := range [][]byte{nil, {}} {
		document, err := lean.OpenMarkdownSource(input)
		if err != nil {
			t.Fatal(err)
		}
		output, err := document.RenderHTML()
		if err != nil || len(output) != 0 || len(document.Bytes()) != 0 {
			t.Fatalf("empty source failed: %s (%v)", output, err)
		}
	}
}

func TestMarkdownSourceHidesUnsafePreviewContent(t *testing.T) {
	for _, name := range []string{"raw-html", "unsafe-links"} {
		source, err := os.ReadFile("testdata/fixtures/markdown/" + name + ".md")
		if err != nil {
			t.Fatal(err)
		}
		document, err := lean.OpenMarkdownSource(source)
		if err != nil {
			t.Fatal(err)
		}
		rendered, err := document.RenderHTML()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(rendered), "<script") || strings.Contains(strings.ToLower(string(rendered)), "javascript:") {
			t.Fatalf("unsafe content was rendered: %s", rendered)
		}
		if !bytes.Equal(document.Bytes(), source) {
			t.Fatal("preview changed original source")
		}
	}
}

func TestMarkdownPreviewSanitizesAttributesAndKeepsImages(t *testing.T) {
	source, err := os.ReadFile("testdata/fixtures/markdown/unsafe-attributes.md")
	if err != nil {
		t.Fatal(err)
	}
	document, err := lean.OpenMarkdownSource(source)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := document.RenderHTML()
	if err != nil {
		t.Fatal(err)
	}
	html := string(rendered)
	for _, blocked := range []string{"style=", "onmouseover", "onmousemove", "onclick"} {
		if strings.Contains(html, blocked) {
			t.Fatalf("unsafe HTML attribute remains: %s", html)
		}
	}
	if !strings.Contains(html, `<img src="http://example.org"`) || !strings.Contains(html, `target="_blank"`) || !strings.Contains(html, "noreferrer") {
		t.Fatalf("image or link isolation missing: %s", html)
	}
	if !bytes.Equal(document.Bytes(), source) {
		t.Fatal("sanitization changed source")
	}
	if _, err := document.RenderSyntaxHTML(lean.MarkdownDialect("unknown")); err == nil {
		t.Fatal("unsupported dialect accepted")
	}
}

func TestMarkdownSourcePreservesNULAndRendersReplacementCharacters(t *testing.T) {
	fixture, err := os.ReadFile("testdata/fixtures/markdown/table.md")
	if err != nil {
		t.Fatal(err)
	}
	source := bytes.Replace(fixture, []byte(" "), []byte{0}, 1)
	document, err := lean.OpenMarkdownSource(source)
	if err != nil {
		t.Fatal(err)
	}
	for _, dialect := range []lean.MarkdownDialect{lean.CommonMark, lean.GitHubFlavoredMarkdown} {
		html, err := document.RenderSyntaxHTML(dialect)
		if err != nil {
			t.Fatal(err)
		}
		if bytes.ContainsRune(html, 0) || !strings.Contains(string(html), "\ufffd") {
			t.Fatalf("NUL was not replaced in rendering: %q", html)
		}
	}
	if !bytes.Equal(source, document.Bytes()) {
		t.Fatal("NUL normalization changed canonical source")
	}
}
