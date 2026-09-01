package md

import (
	"fmt"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	ext "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"gopkg.in/yaml.v3"
)

var markdown = goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithParserOptions(parser.WithAutoHeadingID()))

type converter struct {
	source   []byte
	document *ir.Document
	blockID  int
	listID   int
}

type inlineState struct {
	attrs     ir.RunAttrs
	hyperlink *ir.Hyperlink
}

func Parse(input []byte) (*ir.Document, error) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = nil
	content, err := parseFrontmatter(input, document)
	if err != nil {
		return nil, err
	}
	root := markdown.Parser().Parse(text.NewReader(content))
	state := &converter{source: content, document: document}
	for node := root.FirstChild(); node != nil; node = node.NextSibling() {
		state.convertBlock(node, 0, "")
	}
	return document, nil
}

func parseFrontmatter(input []byte, document *ir.Document) ([]byte, error) {
	text := string(input)
	if !strings.HasPrefix(text, "---\n") {
		return input, nil
	}
	end := strings.Index(text[4:], "\n---")
	if end < 0 {
		return input, nil
	}
	end += 4
	var metadata struct {
		Title  string `yaml:"title"`
		Author string `yaml:"author"`
	}
	if err := yaml.Unmarshal([]byte(text[4:end]), &metadata); err != nil {
		return nil, fmt.Errorf("md: parse frontmatter: %w", err)
	}
	document.Meta.Title = metadata.Title
	document.Meta.Author = metadata.Author
	return []byte(strings.TrimPrefix(text[end+4:], "\n")), nil
}

func (c *converter) nextID(prefix string) string {
	c.blockID++
	return fmt.Sprintf("%s%d", prefix, c.blockID)
}

func (c *converter) convertBlock(node ast.Node, listLevel int, numberingID string) {
	switch value := node.(type) {
	case *ast.Paragraph:
		c.appendParagraph(value, "", listLevel, numberingID, 0)
	case *ast.TextBlock:
		c.appendParagraph(value, "", listLevel, numberingID, 0)
	case *ast.Heading:
		c.appendParagraph(value, fmt.Sprintf("Heading%d", value.Level), listLevel, numberingID, 0)
	case *ast.List:
		c.convertList(value, listLevel, numberingID)
	case *ast.Blockquote:
		for child := value.FirstChild(); child != nil; child = child.NextSibling() {
			c.convertQuotedBlock(child)
		}
	case *ast.FencedCodeBlock:
		c.appendCodeBlock(value.Lines().Value(c.source))
	case *ast.CodeBlock:
		c.appendCodeBlock(value.Lines().Value(c.source))
	case *ast.ThematicBreak:
		c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, &ir.Paragraph{ID: c.nextID("p"), Para: ir.ParaAttrs{Borders: ir.ParaBorders{Bottom: ir.Border{Style: "single", Width: 1}}}})
	case *ext.Table:
		c.convertTable(value)
	}
}

func (c *converter) appendParagraph(node ast.Node, style string, listLevel int, numberingID string, indent float64) {
	runs := normalizeRuns(c.convertInlines(node, inlineState{}))
	if len(runs) == 0 {
		return
	}
	paragraph := &ir.Paragraph{ID: c.nextID("p"), Style: style, Runs: runs}
	paragraph.Para.Indent.Left = indent
	if numberingID != "" {
		paragraph.Numbering = &ir.NumberingRef{ID: numberingID, Level: listLevel}
	}
	c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, paragraph)
}

func (c *converter) convertQuotedBlock(node ast.Node) {
	switch value := node.(type) {
	case *ast.Paragraph:
		c.appendParagraph(value, "", 0, "", 36)
	case *ast.TextBlock:
		c.appendParagraph(value, "", 0, "", 36)
	default:
		c.convertBlock(node, 0, "")
	}
}

func (c *converter) appendCodeBlock(content []byte) {
	c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, &ir.Paragraph{ID: c.nextID("p"), Runs: []ir.Run{{Text: strings.TrimSuffix(string(content), "\n"), Attrs: ir.RunAttrs{FontName: "monospace"}}}})
}

func (c *converter) convertList(list *ast.List, level int, numberingID string) {
	if numberingID == "" {
		c.listID++
		numberingID = fmt.Sprintf("md-list-%d", c.listID)
		format := ir.NumFormatBullet
		if list.IsOrdered() {
			format = ir.NumFormatDecimal
		}
		levels := make([]ir.NumberingLevel, 9)
		for index := range levels {
			levels[index] = ir.NumberingLevel{Level: index, Format: format, Start: 1}
		}
		c.document.Styles.Numbering = append(c.document.Styles.Numbering, ir.NumberingDef{ID: numberingID, Levels: levels})
	}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		for child := item.FirstChild(); child != nil; child = child.NextSibling() {
			if nested, ok := child.(*ast.List); ok {
				c.convertList(nested, level+1, numberingID)
			} else {
				c.convertBlock(child, level, numberingID)
			}
		}
	}
}

func (c *converter) convertTable(table *ext.Table) {
	result := &ir.Table{ID: c.nextID("t")}
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		row := ir.TableRow{}
		_, row.IsHeader = child.(*ext.TableHeader)
		for cellNode := child.FirstChild(); cellNode != nil; cellNode = cellNode.NextSibling() {
			cell, ok := cellNode.(*ext.TableCell)
			if !ok {
				continue
			}
			paragraph := &ir.Paragraph{ID: c.nextID("p"), Runs: normalizeRuns(c.convertInlines(cell, inlineState{}))}
			paragraph.Para.Align = tableAlignment(cell.Alignment)
			row.Cells = append(row.Cells, ir.TableCell{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{paragraph}})
		}
		result.Rows = append(result.Rows, row)
	}
	c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, result)
}

func tableAlignment(alignment ext.Alignment) ir.Alignment {
	switch alignment {
	case ext.AlignCenter:
		return ir.AlignCenter
	case ext.AlignRight:
		return ir.AlignRight
	default:
		return ir.AlignLeft
	}
}

func (c *converter) convertInlines(parent ast.Node, state inlineState) []ir.Run {
	var runs []ir.Run
	for node := parent.FirstChild(); node != nil; node = node.NextSibling() {
		switch value := node.(type) {
		case *ast.Text:
			runs = append(runs, ir.Run{Text: string(value.Segment.Value(c.source)), Attrs: state.attrs, Hyperlink: cloneHyperlink(state.hyperlink)})
			if value.HardLineBreak() {
				runs = append(runs, ir.Run{Break: ir.BreakLine, Attrs: state.attrs, Hyperlink: cloneHyperlink(state.hyperlink)})
			} else if value.SoftLineBreak() {
				runs = append(runs, ir.Run{Text: " ", Attrs: state.attrs, Hyperlink: cloneHyperlink(state.hyperlink)})
			}
		case *ast.String:
			runs = append(runs, ir.Run{Text: string(value.Value), Attrs: state.attrs, Hyperlink: cloneHyperlink(state.hyperlink)})
		case *ast.Emphasis:
			next := state
			if value.Level == 2 {
				next.attrs.Bold = true
			} else {
				next.attrs.Italic = true
			}
			runs = append(runs, c.convertInlines(value, next)...)
		case *ext.Strikethrough:
			next := state
			next.attrs.Strike = true
			runs = append(runs, c.convertInlines(value, next)...)
		case *ast.CodeSpan:
			next := state
			next.attrs.FontName = "monospace"
			runs = append(runs, c.convertInlines(value, next)...)
		case *ast.Link:
			next := state
			next.hyperlink = &ir.Hyperlink{URL: string(value.Destination)}
			runs = append(runs, c.convertInlines(value, next)...)
		case *ast.Image:
			c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, &ir.Image{ID: c.nextID("image"), Alt: inlineText(value, c.source)})
		default:
			runs = append(runs, c.convertInlines(value, state)...)
		}
	}
	return runs
}

func inlineText(parent ast.Node, source []byte) string {
	var text strings.Builder
	_ = ast.Walk(parent, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			switch value := node.(type) {
			case *ast.Text:
				text.Write(value.Segment.Value(source))
			case *ast.String:
				text.Write(value.Value)
			}
		}
		return ast.WalkContinue, nil
	})
	return text.String()
}

func cloneHyperlink(link *ir.Hyperlink) *ir.Hyperlink {
	if link == nil {
		return nil
	}
	copy := *link
	return &copy
}

func normalizeRuns(runs []ir.Run) []ir.Run {
	result := make([]ir.Run, 0, len(runs))
	for _, run := range runs {
		if len(result) > 0 && run.Break == ir.BreakNone && result[len(result)-1].Break == ir.BreakNone && run.Attrs == result[len(result)-1].Attrs && equalHyperlink(run.Hyperlink, result[len(result)-1].Hyperlink) {
			result[len(result)-1].Text += run.Text
			continue
		}
		result = append(result, run)
	}
	return result
}

func equalHyperlink(left, right *ir.Hyperlink) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
