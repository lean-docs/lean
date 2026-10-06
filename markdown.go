package lean

import (
	"bytes"
	"errors"
	"unicode/utf8"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

var ErrInvalidMarkdown = errors.New("lean: Markdown must contain valid UTF-8 text without NUL characters")

type MarkdownSource struct {
	source []byte
}

func OpenMarkdownSource(source []byte) (*MarkdownSource, error) {
	if !utf8.Valid(source) || bytes.ContainsRune(source, 0) {
		return nil, ErrInvalidMarkdown
	}
	return &MarkdownSource{source: bytes.Clone(source)}, nil
}

func (document *MarkdownSource) Bytes() []byte {
	return bytes.Clone(document.source)
}

func (document *MarkdownSource) RenderHTML() ([]byte, error) {
	engine := goldmark.New(goldmark.WithExtensions(extension.GFM))
	content := bytes.TrimPrefix(document.source, []byte{0xef, 0xbb, 0xbf})
	root := engine.Parser().Parse(text.NewReader(content))
	if err := ast.Walk(root, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && (node.Kind() == ast.KindLink || node.Kind() == ast.KindAutoLink) {
			node.SetAttributeString("target", []byte("_blank"))
			node.SetAttributeString("rel", []byte("noopener noreferrer"))
		}
		return ast.WalkContinue, nil
	}); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if err := engine.Renderer().Render(&output, content, root); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
