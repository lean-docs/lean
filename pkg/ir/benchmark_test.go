package ir_test

import (
	"fmt"
	"testing"

	"github.com/lean-docs/lean/pkg/ir"
)

// --- Cluster 14: Performance Baselines ---

// C14.1 BenchmarkParseMD10Page measures parsing a 10-page Markdown document into IR.
func BenchmarkParseMD10Page(b *testing.B) {
	b.Skip("benchmark target not implemented: markdown parser")

	// md := generateMarkdown(10)
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	_, err := parser.ParseMarkdown([]byte(md))
	// 	if err != nil {
	// 		b.Fatal(err)
	// 	}
	// }
}

// C14.2 BenchmarkParseDocx10Page measures parsing a 10-page DOCX into IR.
func BenchmarkParseDocx10Page(b *testing.B) {
	b.Skip("benchmark target not implemented: docx parser")

	// docx, err := os.ReadFile("testdata/10page.docx")
	// require.NoError(b, err)
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	_, err := parser.ParseDocx(docx)
	// 	if err != nil {
	// 		b.Fatal(err)
	// 	}
	// }
}

// C14.3 BenchmarkExportTypst10Page measures exporting a 10-page IR document to Typst.
func BenchmarkExportTypst10Page(b *testing.B) {
	b.Skip("benchmark target not implemented: typst exporter")

	// doc := build10PageDoc()
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	_, err := export.ToTypst(doc)
	// 	if err != nil {
	// 		b.Fatal(err)
	// 	}
	// }
}

// C14.4 BenchmarkExportDocx10Page measures exporting a 10-page IR document to DOCX.
func BenchmarkExportDocx10Page(b *testing.B) {
	b.Skip("benchmark target not implemented: docx exporter")

	// doc := build10PageDoc()
	// b.ResetTimer()
	// for i := 0; i < b.N; i++ {
	// 	_, err := export.ToDocx(doc)
	// 	if err != nil {
	// 		b.Fatal(err)
	// 	}
	// }
}

// C14.5 BenchmarkMemory10Page measures memory consumption for a 10-page IR document.
// This benchmark can run now since it only depends on the IR package.
func BenchmarkMemory10Page(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = build10PageDoc()
	}
}

// build10PageDoc constructs a realistic 10-page IR document.
// Assumes ~40 paragraphs per page, each with 2-3 runs.
func build10PageDoc() *ir.Document {
	doc := ir.NewDocument()
	doc.Meta.Title = "Benchmark Document"
	doc.Meta.Author = "lean-bench"

	const pagesCount = 10
	const parasPerPage = 40

	blocks := make([]ir.Block, 0, pagesCount*parasPerPage)
	for page := 0; page < pagesCount; page++ {
		for para := 0; para < parasPerPage; para++ {
			idx := page*parasPerPage + para
			p := &ir.Paragraph{
				ID:    fmt.Sprintf("p%d", idx),
				Style: "Normal",
				Runs: []ir.Run{
					{
						Text: fmt.Sprintf("This is paragraph %d on page %d with enough text to simulate real content. ", para, page),
						Attrs: ir.RunAttrs{
							FontName: "Calibri",
							FontSize: 11,
						},
					},
					{
						Text:  "Bold segment. ",
						Attrs: ir.RunAttrs{Bold: true, FontName: "Calibri", FontSize: 11},
					},
					{
						Text:  "Italic ending.",
						Attrs: ir.RunAttrs{Italic: true, FontName: "Calibri", FontSize: 11},
					},
				},
				Para: ir.ParaAttrs{
					Spacing: ir.Spacing{After: 6, Line: 1.15, LineRule: ir.LineRuleAuto},
				},
			}
			// Every 10th paragraph is a heading
			if para == 0 {
				p.Style = "Heading1"
				p.Runs = []ir.Run{{
					Text:  fmt.Sprintf("Chapter %d", page+1),
					Attrs: ir.RunAttrs{Bold: true, FontName: "Calibri", FontSize: 24},
				}}
				p.Para.Spacing = ir.Spacing{Before: 24, After: 12}
			}
			blocks = append(blocks, p)
		}

		// Add a page break paragraph between pages (except after the last)
		if page < pagesCount-1 {
			blocks = append(blocks, &ir.Paragraph{
				ID:   fmt.Sprintf("pb%d", page),
				Runs: []ir.Run{{Break: ir.BreakPage}},
			})
		}
	}
	doc.Sections[0].Blocks = blocks
	return doc
}
