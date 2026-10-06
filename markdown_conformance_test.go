package lean_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	lean "github.com/lean-docs/lean"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func TestMarkdownSpecificationCorpora(t *testing.T) {
	for _, profile := range []struct {
		name    string
		dialect lean.MarkdownDialect
		count   int
	}{
		{"commonmark-0.31.2", lean.CommonMark, 652},
		{"gfm-0.29-extensions", lean.GitHubFlavoredMarkdown, 24},
	} {
		t.Run(profile.name, func(t *testing.T) {
			raw, err := os.ReadFile("testdata/fixtures/markdown/conformance/" + profile.name + ".json")
			if err != nil {
				t.Fatal(err)
			}
			var examples []struct {
				Markdown, HTML, Section string
				Example                 int
			}
			if err := json.Unmarshal(raw, &examples); err != nil {
				t.Fatal(err)
			}
			if len(examples) != profile.count {
				t.Fatalf("corpus size changed: %d", len(examples))
			}
			for _, example := range examples {
				t.Run(fmt.Sprintf("%s/%d", example.Section, example.Example), func(t *testing.T) {
					document, err := lean.OpenMarkdownSource([]byte(example.Markdown))
					if err != nil {
						t.Fatal(err)
					}
					actual, err := document.RenderSyntaxHTML(profile.dialect)
					if err != nil {
						t.Fatal(err)
					}
					if profile.dialect == lean.CommonMark && string(actual) != example.HTML || profile.dialect == lean.GitHubFlavoredMarkdown && normalizeHTML(t, string(actual)) != normalizeHTML(t, example.HTML) {
						t.Fatalf("expected %q, got %q", example.HTML, actual)
					}
					if string(document.Bytes()) != example.Markdown {
						t.Fatal("rendering changed canonical source")
					}
					safe, err := document.RenderHTML()
					if err != nil {
						t.Fatal(err)
					}
					nodes, err := html.ParseFragment(strings.NewReader(string(safe)), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
					if err != nil {
						t.Fatal(err)
					}
					var inspect func(*html.Node)
					inspect = func(node *html.Node) {
						if node.Type == html.ElementNode {
							switch node.Data {
							case "script", "style", "iframe", "object", "embed":
								t.Errorf("unsafe element %s", node.Data)
							}
							for _, attr := range node.Attr {
								if strings.HasPrefix(strings.ToLower(attr.Key), "on") || attr.Key == "style" {
									t.Errorf("unsafe attribute %s", attr.Key)
								}
							}
						}
						for child := node.FirstChild; child != nil; child = child.NextSibling {
							inspect(child)
						}
					}
					for _, node := range nodes {
						inspect(node)
					}
				})
			}
		})
	}
}

func normalizeHTML(t *testing.T, source string) string {
	t.Helper()
	nodes, err := html.ParseFragment(strings.NewReader(source), &html.Node{Type: html.ElementNode, Data: "div", DataAtom: atom.Div})
	if err != nil {
		t.Fatal(err)
	}
	var sortAttributes func(*html.Node)
	sortAttributes = func(node *html.Node) {
		sort.Slice(node.Attr, func(i, j int) bool { return node.Attr[i].Key < node.Attr[j].Key })
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			sortAttributes(c)
		}
	}
	var result bytes.Buffer
	for _, node := range nodes {
		sortAttributes(node)
		if err := html.Render(&result, node); err != nil {
			t.Fatal(err)
		}
	}
	return result.String()
}
