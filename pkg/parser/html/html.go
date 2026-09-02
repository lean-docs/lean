package html

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
	xhtml "golang.org/x/net/html"
)

type converter struct {
	document *ir.Document
	blockID  int
	listID   int
}

type inlineState struct {
	attrs     ir.RunAttrs
	hyperlink *ir.Hyperlink
}

func Parse(input []byte) (*ir.Document, error) {
	root, err := xhtml.Parse(bytes.NewReader(input))
	if err != nil {
		return nil, fmt.Errorf("html: parse document: %w", err)
	}
	document := ir.NewDocument()
	document.Sections[0].Blocks = nil
	state := &converter{document: document}
	body := findElement(root, "body")
	if body == nil {
		body = root
	}
	for node := body.FirstChild; node != nil; node = node.NextSibling {
		state.convertBlock(node, 0, "", 0)
	}
	return document, nil
}

func (c *converter) nextID(prefix string) string {
	c.blockID++
	return fmt.Sprintf("%s%d", prefix, c.blockID)
}

func (c *converter) convertBlock(node *xhtml.Node, listLevel int, numberingID string, indent float64) {
	if node.Type == xhtml.TextNode {
		if strings.TrimSpace(node.Data) != "" {
			c.appendParagraph(node, "", listLevel, numberingID, indent)
		}
		return
	}
	if node.Type != xhtml.ElementNode {
		return
	}
	switch node.Data {
	case "p":
		c.appendParagraph(node, "", listLevel, numberingID, indent)
	case "h1", "h2", "h3", "h4", "h5", "h6":
		level, _ := strconv.Atoi(node.Data[1:])
		c.appendHeading(node, level)
	case "img":
		c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, c.convertImage(node))
	case "ul", "ol":
		c.convertList(node, listLevel, numberingID)
	case "blockquote":
		if hasBlockChild(node) {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				c.convertBlock(child, 0, "", indent+36)
			}
		} else {
			c.appendParagraph(node, "", 0, "", indent+36)
		}
	case "pre":
		c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, &ir.Paragraph{ID: c.nextID("p"), Runs: []ir.Run{{Text: nodeText(node), Attrs: ir.RunAttrs{FontName: "monospace"}}}})
	case "hr":
		c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, &ir.Paragraph{ID: c.nextID("p"), Para: ir.ParaAttrs{Borders: ir.ParaBorders{Bottom: ir.Border{Style: "single", Width: 1}}}})
	case "table":
		c.convertTable(node)
	default:
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			c.convertBlock(child, listLevel, numberingID, indent)
		}
	}
}

func (c *converter) appendParagraph(node *xhtml.Node, style string, listLevel int, numberingID string, indent float64) {
	runs := normalizeRuns(c.convertInlineChildren(node, inlineState{}))
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

func (c *converter) appendHeading(node *xhtml.Node, level int) {
	runs := normalizeRuns(c.convertInlineChildren(node, inlineState{}))
	c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, &ir.Paragraph{ID: c.nextID("p"), Style: fmt.Sprintf("Heading%d", level), Runs: runs, Para: ir.ParaAttrs{OutlineLevel: level}})
}

func (c *converter) convertList(list *xhtml.Node, level int, numberingID string) {
	if numberingID == "" {
		c.listID++
		numberingID = fmt.Sprintf("html-list-%d", c.listID)
		format := ir.NumFormatBullet
		if list.Data == "ol" {
			format = ir.NumFormatDecimal
		}
		levels := make([]ir.NumberingLevel, 9)
		for index := range levels {
			levels[index] = ir.NumberingLevel{Level: index, Format: format, Start: 1}
		}
		c.document.Styles.Numbering = append(c.document.Styles.Numbering, ir.NumberingDef{ID: numberingID, Levels: levels})
	}
	for item := list.FirstChild; item != nil; item = item.NextSibling {
		if item.Type != xhtml.ElementNode || item.Data != "li" {
			continue
		}
		paragraph := &ir.Paragraph{ID: c.nextID("p"), Runs: normalizeRuns(c.convertListItemText(item)), Numbering: &ir.NumberingRef{ID: numberingID, Level: level}}
		c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, paragraph)
		for child := item.FirstChild; child != nil; child = child.NextSibling {
			if child.Type == xhtml.ElementNode && (child.Data == "ul" || child.Data == "ol") {
				c.convertList(child, level+1, numberingID)
			}
		}
	}
}

func (c *converter) convertListItemText(item *xhtml.Node) []ir.Run {
	var runs []ir.Run
	for child := item.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xhtml.ElementNode && (child.Data == "ul" || child.Data == "ol") {
			continue
		}
		runs = append(runs, c.convertInline(child, inlineState{})...)
	}
	return runs
}

func (c *converter) convertTable(table *xhtml.Node) {
	result := &ir.Table{ID: c.nextID("t")}
	var walkRows func(*xhtml.Node, bool)
	walkRows = func(parent *xhtml.Node, inHeader bool) {
		for node := parent.FirstChild; node != nil; node = node.NextSibling {
			if node.Type != xhtml.ElementNode {
				continue
			}
			switch node.Data {
			case "thead":
				walkRows(node, true)
			case "tbody", "tfoot":
				walkRows(node, false)
			case "tr":
				row := ir.TableRow{IsHeader: inHeader}
				for cellNode := node.FirstChild; cellNode != nil; cellNode = cellNode.NextSibling {
					if cellNode.Type != xhtml.ElementNode || (cellNode.Data != "td" && cellNode.Data != "th") {
						continue
					}
					row.IsHeader = row.IsHeader || cellNode.Data == "th"
					span := attributeInt(cellNode, "colspan", 1)
					paragraph := &ir.Paragraph{ID: c.nextID("p"), Runs: normalizeRuns(c.convertInlineChildren(cellNode, inlineState{}))}
					row.Cells = append(row.Cells, ir.TableCell{ColSpan: span, RowSpan: 1, Blocks: []ir.Block{paragraph}})
				}
				result.Rows = append(result.Rows, row)
			}
		}
	}
	walkRows(table, false)
	c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, result)
}

func (c *converter) convertInlineChildren(parent *xhtml.Node, state inlineState) []ir.Run {
	var runs []ir.Run
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		runs = append(runs, c.convertInline(child, state)...)
	}
	return runs
}

func (c *converter) convertInline(node *xhtml.Node, state inlineState) []ir.Run {
	if node.Type == xhtml.TextNode {
		return []ir.Run{{Text: node.Data, Attrs: state.attrs, Hyperlink: cloneHyperlink(state.hyperlink)}}
	}
	if node.Type != xhtml.ElementNode {
		return nil
	}
	next := state
	switch node.Data {
	case "strong", "b":
		next.attrs.Bold = true
	case "em", "i":
		next.attrs.Italic = true
	case "u":
		next.attrs.Underline = ir.UnderlineSingle
	case "s", "del":
		next.attrs.Strike = true
	case "code":
		next.attrs.FontName = "monospace"
	case "a":
		next.hyperlink = &ir.Hyperlink{URL: attribute(node, "href")}
	case "br":
		return []ir.Run{{Break: ir.BreakLine, Attrs: next.attrs, Hyperlink: cloneHyperlink(next.hyperlink)}}
	case "img":
		c.document.Sections[0].Blocks = append(c.document.Sections[0].Blocks, c.convertImage(node))
		return nil
	}
	applyInlineStyle(&next.attrs, attribute(node, "style"))
	return c.convertInlineChildren(node, next)
}

func (c *converter) convertImage(node *xhtml.Node) *ir.Image {
	return &ir.Image{ID: c.nextID("image"), Alt: attribute(node, "alt")}
}

func applyInlineStyle(attrs *ir.RunAttrs, style string) {
	for _, declaration := range strings.Split(style, ";") {
		parts := strings.SplitN(declaration, ":", 2)
		if len(parts) != 2 {
			continue
		}
		property := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(strings.ToLower(parts[1]))
		switch property {
		case "font-weight":
			attrs.Bold = value == "bold" || value == "bolder" || numericWeight(value) >= 600
		case "font-style":
			attrs.Italic = value == "italic" || value == "oblique"
		case "color":
			if color, ok := parseColor(value); ok {
				attrs.Color = color
			}
		case "font-size":
			if strings.HasSuffix(value, "pt") {
				attrs.FontSize, _ = strconv.ParseFloat(strings.TrimSuffix(value, "pt"), 64)
			}
		}
	}
}

func parseColor(value string) (ir.Color, bool) {
	value = strings.TrimPrefix(value, "#")
	if len(value) != 6 {
		return ir.Color{}, false
	}
	number, err := strconv.ParseUint(value, 16, 32)
	if err != nil {
		return ir.Color{}, false
	}
	return ir.Color{R: uint8(number >> 16), G: uint8(number >> 8), B: uint8(number)}, true
}

func numericWeight(value string) int {
	weight, _ := strconv.Atoi(value)
	return weight
}

func normalizeRuns(runs []ir.Run) []ir.Run {
	result := make([]ir.Run, 0, len(runs))
	for _, run := range runs {
		if len(result) > 0 && run.Break == ir.BreakNone && result[len(result)-1].Break == ir.BreakNone && run.Attrs == result[len(result)-1].Attrs && equalHyperlink(run.Hyperlink, result[len(result)-1].Hyperlink) {
			result[len(result)-1].Text += run.Text
		} else {
			result = append(result, run)
		}
	}
	return result
}

func findElement(node *xhtml.Node, name string) *xhtml.Node {
	if node.Type == xhtml.ElementNode && node.Data == name {
		return node
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if match := findElement(child, name); match != nil {
			return match
		}
	}
	return nil
}

func hasBlockChild(node *xhtml.Node) bool {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == xhtml.ElementNode && (child.Data == "p" || child.Data == "div" || child.Data == "ul" || child.Data == "ol" || child.Data == "table") {
			return true
		}
	}
	return false
}

func nodeText(node *xhtml.Node) string {
	if node.Type == xhtml.TextNode {
		return node.Data
	}
	var text strings.Builder
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text.WriteString(nodeText(child))
	}
	return text.String()
}

func attribute(node *xhtml.Node, name string) string {
	for _, value := range node.Attr {
		if value.Key == name {
			return value.Val
		}
	}
	return ""
}

func attributeInt(node *xhtml.Node, name string, fallback int) int {
	value, err := strconv.Atoi(attribute(node, name))
	if err != nil || value < 1 {
		return fallback
	}
	return value
}

func cloneHyperlink(link *ir.Hyperlink) *ir.Hyperlink {
	if link == nil {
		return nil
	}
	copy := *link
	return &copy
}

func equalHyperlink(left, right *ir.Hyperlink) bool {
	if left == nil || right == nil {
		return left == right
	}
	return *left == *right
}
