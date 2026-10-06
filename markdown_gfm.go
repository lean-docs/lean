package lean

import (
	"bytes"
	"regexp"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var disallowedMarkdownTags = regexp.MustCompile(`(?i)<(/?)(title|textarea|style|xmp|iframe|noembed|noframes|script|plaintext)([\t\n\f\r >]|/>)`)
var disallowedMarkdownTagStart = regexp.MustCompile(`(?i)^<(/?)(title|textarea|style|xmp|iframe|noembed|noframes|script|plaintext)([\t\n\f\r >]|/>)`)

type markdownTagFilter struct{}

func (markdownTagFilter) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindRawHTML, renderMarkdownRawHTML)
	reg.Register(ast.KindHTMLBlock, renderMarkdownHTMLBlock)
}

func renderMarkdownRawHTML(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkSkipChildren, nil
	}
	raw := node.(*ast.RawHTML)
	var content bytes.Buffer
	for index := 0; index < raw.Segments.Len(); index++ {
		segment := raw.Segments.At(index)
		content.Write(segment.Value(source))
	}
	value := content.Bytes()
	if disallowedMarkdownTagStart.Match(value) {
		value = append([]byte("&lt;"), value[1:]...)
	}
	_, err := writer.Write(value)
	return ast.WalkSkipChildren, err
}

func renderMarkdownHTMLBlock(writer util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	block := node.(*ast.HTMLBlock)
	if entering {
		for index := 0; index < block.Lines().Len(); index++ {
			segment := block.Lines().At(index)
			if _, err := writer.Write(disallowedMarkdownTags.ReplaceAll(segment.Value(source), []byte("&lt;$1$2$3"))); err != nil {
				return ast.WalkStop, err
			}
		}
	} else if block.HasClosure() {
		if _, err := writer.Write(disallowedMarkdownTags.ReplaceAll(block.ClosureLine.Value(source), []byte("&lt;$1$2$3"))); err != nil {
			return ast.WalkStop, err
		}
	}
	return ast.WalkContinue, nil
}
