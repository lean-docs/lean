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
	for _, input := range [][]byte{{0xff}, {0}, {'a', 0, 'b'}} {
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
		if name == "unsafe-links" && (!strings.Contains(string(rendered), `target="_blank"`) || !strings.Contains(string(rendered), `rel="noopener noreferrer"`)) {
			t.Fatalf("links lack navigation isolation: %s", rendered)
		}
	}
}
