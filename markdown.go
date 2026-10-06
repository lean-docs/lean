package lean

import (
	"bytes"
	"errors"
	nethtml "golang.org/x/net/html"
	"io"
	"regexp"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/util"
)

var ErrInvalidMarkdown = errors.New("lean: Markdown must contain valid UTF-8 text")

type MarkdownSource struct {
	source []byte
}

func OpenMarkdownSource(source []byte) (*MarkdownSource, error) {
	if !utf8.Valid(source) {
		return nil, ErrInvalidMarkdown
	}
	return &MarkdownSource{source: bytes.Clone(source)}, nil
}

func (document *MarkdownSource) Bytes() []byte {
	return bytes.Clone(document.source)
}

type MarkdownDialect string

const (
	CommonMark             MarkdownDialect = "commonmark"
	GitHubFlavoredMarkdown MarkdownDialect = "gfm"
)

func (document *MarkdownSource) RenderSyntaxHTML(dialect MarkdownDialect) ([]byte, error) {
	return document.renderSyntaxHTML(dialect, true)
}

func (document *MarkdownSource) renderSyntaxHTML(dialect MarkdownDialect, xhtml bool) ([]byte, error) {
	options := []goldmark.Option{goldmark.WithRendererOptions(html.WithUnsafe())}
	if xhtml {
		options = append(options, goldmark.WithRendererOptions(html.WithXHTML()))
	}
	switch dialect {
	case CommonMark:
	case GitHubFlavoredMarkdown:
		options = append(options, goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(renderer.WithNodeRenderers(util.Prioritized(markdownTagFilter{}, 100))))
	default:
		return nil, errors.New("lean: unsupported Markdown dialect")
	}
	engine := goldmark.New(options...)
	content := bytes.TrimPrefix(document.source, []byte{0xef, 0xbb, 0xbf})
	content = bytes.ReplaceAll(content, []byte{0}, []byte("\ufffd"))
	var output bytes.Buffer
	if err := engine.Convert(content, &output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}

func (document *MarkdownSource) RenderHTML() ([]byte, error) {
	output, err := document.renderSyntaxHTML(GitHubFlavoredMarkdown, false)
	if err != nil {
		return nil, err
	}
	policy := bluemonday.UGCPolicy()
	policy.RequireNoFollowOnLinks(false)
	policy.RequireNoReferrerOnLinks(true)
	policy.AddTargetBlankToFullyQualifiedLinks(true)
	policy.AllowDataURIImages()
	policy.AllowAttrs("align").Matching(regexp.MustCompile(`(?i)^(left|right|center|justify)$`)).OnElements("p", "div", "th", "td")
	policy.AllowAttrs("class").Matching(regexp.MustCompile(`^language-[a-zA-Z0-9_+-]+$`)).OnElements("code")
	policy.AllowAttrs("type").Matching(regexp.MustCompile(`^checkbox$`)).OnElements("input")
	policy.AllowAttrs("disabled", "checked").Matching(regexp.MustCompile(`^(|disabled|checked)$`)).OnElements("input")
	policy.AllowElements("input")
	sanitized := policy.SanitizeBytes(output)
	return isolateMarkdownLinksAndInputs(sanitized), nil
}

func isolateMarkdownLinksAndInputs(source []byte) []byte {
	tokenizer := nethtml.NewTokenizer(bytes.NewReader(source))
	var output bytes.Buffer
	for {
		kind := tokenizer.Next()
		if kind == nethtml.ErrorToken {
			if tokenizer.Err() != io.EOF {
				return source
			}
			return output.Bytes()
		}
		raw := bytes.Clone(tokenizer.Raw())
		if kind != nethtml.StartTagToken && kind != nethtml.SelfClosingTagToken {
			output.Write(raw)
			continue
		}
		token := tokenizer.Token()
		if token.Data != "a" && token.Data != "input" {
			output.Write(raw)
			continue
		}
		set := func(key, value string) {
			for index := range token.Attr {
				if token.Attr[index].Key == key {
					token.Attr[index].Val = value
					return
				}
			}
			token.Attr = append(token.Attr, nethtml.Attribute{Key: key, Val: value})
		}
		if token.Data == "input" {
			checkbox := false
			for _, attr := range token.Attr {
				if attr.Key == "type" && attr.Val == "checkbox" {
					checkbox = true
				}
			}
			if !checkbox {
				continue
			}
			set("disabled", "")
		} else {
			set("target", "_blank")
			set("rel", "noopener noreferrer")
		}
		output.WriteString(token.String())
	}
}
