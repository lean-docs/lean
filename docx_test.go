package lean_test

import (
	"os"
	"testing"

	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenDOCXReportsSupportedCorpusFeatures(t *testing.T) {
	content, err := os.ReadFile("testdata/fixtures/ooxml/python-docx/doc-coreprops.docx")
	require.NoError(t, err)
	document, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.NotNil(t, document)
	assert.True(t, report.Editable)
	assert.Empty(t, report.Unsupported)
	assert.NotEmpty(t, document.Meta.Title)
	assert.NotEmpty(t, document.Meta.Author)
}

func TestOpenDOCXPreservesImageWhileReportingOtherUnsupportedFeatures(t *testing.T) {
	content, err := os.ReadFile("testdata/fixtures/ooxml/python-docx/having-images.docx")
	require.NoError(t, err)
	document, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.False(t, report.Editable)
	assert.NotContains(t, report.Unsupported, "images")
	assert.Contains(t, report.Unsupported, "headers")
	assert.Contains(t, report.Unsupported, "footnotes")
	assert.True(t, documentHasImage(document))
}

func TestSaveDOCXRoundTripsSupportedDocument(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "A formatted plan", Attrs: ir.RunAttrs{Bold: true}}}},
		&ir.Table{ID: "t1", Rows: []ir.TableRow{{Cells: []ir.TableCell{{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{&ir.Paragraph{ID: "p2", Runs: []ir.Run{{Text: "Milestone"}}}}}}}}},
	}
	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	assert.True(t, report.Editable)
	if directory := os.Getenv("LEAN_VALIDATION_DIR"); directory != "" {
		require.NoError(t, os.WriteFile(directory+"/lean-mvp.docx", content, 0o600))
	}
	reopened, reopenedReport, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.True(t, reopenedReport.Editable)
	require.Len(t, reopened.Sections[0].Blocks, 2)
	paragraph := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "A formatted plan", paragraph.Runs[0].Text)
	assert.True(t, paragraph.Runs[0].Attrs.Bold)
}

func TestSaveDOCXRefusesUnsupportedFeatures(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Image{ID: "image", Data: []byte("image"), Format: ir.ImagePNG}}
	content, report, err := lean.SaveDOCX(document)
	assert.ErrorIs(t, err, lean.ErrUnsupportedDocument)
	assert.Nil(t, content)
	assert.Contains(t, report.Unsupported, "images")
}

func TestSaveDOCXRoundTripsMVPFormatting(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Properties.Orientation = ir.Landscape
	document.Sections[0].Properties.Width = 792
	document.Sections[0].Properties.Height = 612
	document.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Para: ir.ParaAttrs{
				OutlineLevel: 2,
				Borders:      ir.ParaBorders{Bottom: ir.Border{Style: "single", Width: 1, Color: ir.Color{R: 25, G: 50, B: 75}}},
				Shading:      ir.Color{R: 240, G: 241, B: 242},
				BiDi:         true,
			},
			Runs: []ir.Run{{
				Text: "Formatted",
				Attrs: ir.RunAttrs{
					Highlight: ir.Color{R: 255, G: 255, B: 0},
					Tracking:  1.5,
					Language:  "fr-CM",
				},
			}},
		},
		&ir.Table{
			ID:      "t1",
			Align:   ir.AlignCenter,
			Borders: ir.TableBorders{InsideH: ir.Border{Style: "single", Width: 0.5, Color: ir.Color{R: 10, G: 20, B: 30}}},
			Shading: ir.Color{R: 245, G: 246, B: 247},
			Rows: []ir.TableRow{{
				IsHeader: true,
				Height:   floatPointer(24),
				Cells: []ir.TableCell{{
					ColSpan: 1,
					RowSpan: 1,
					VAlign:  ir.VAlignCenter,
					Borders: ir.CellBorders{Left: ir.Border{Style: "single", Width: 0.75, Color: ir.Color{R: 40, G: 50, B: 60}}},
					Shading: ir.Color{R: 230, G: 231, B: 232},
					Padding: ir.Padding{Top: 4, Right: 5, Bottom: 6, Left: 7},
					Blocks:  []ir.Block{&ir.Paragraph{ID: "p2", Runs: []ir.Run{{Text: "Cell"}}}},
				}},
			}},
		},
	}

	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	assert.True(t, report.Editable)
	reopened, reopenedReport, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.True(t, reopenedReport.Editable)
	assert.Empty(t, reopenedReport.Unsupported)
	assert.Equal(t, ir.Landscape, reopened.Sections[0].Properties.Orientation)

	paragraph := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, 2, paragraph.Para.OutlineLevel)
	assert.Equal(t, document.Sections[0].Blocks[0].(*ir.Paragraph).Para.Borders, paragraph.Para.Borders)
	assert.Equal(t, ir.Color{R: 240, G: 241, B: 242}, paragraph.Para.Shading)
	assert.True(t, paragraph.Para.BiDi)
	assert.Equal(t, ir.Color{R: 255, G: 255, B: 0}, paragraph.Runs[0].Attrs.Highlight)
	assert.Equal(t, 1.5, paragraph.Runs[0].Attrs.Tracking)
	assert.Equal(t, "fr-CM", paragraph.Runs[0].Attrs.Language)

	table := reopened.Sections[0].Blocks[1].(*ir.Table)
	assert.Equal(t, ir.AlignCenter, table.Align)
	assert.Equal(t, document.Sections[0].Blocks[1].(*ir.Table).Borders, table.Borders)
	assert.Equal(t, ir.Color{R: 245, G: 246, B: 247}, table.Shading)
	assert.True(t, table.Rows[0].IsHeader)
	require.NotNil(t, table.Rows[0].Height)
	assert.Equal(t, 24.0, *table.Rows[0].Height)
	cell := table.Rows[0].Cells[0]
	assert.Equal(t, ir.VAlignCenter, cell.VAlign)
	assert.Equal(t, document.Sections[0].Blocks[1].(*ir.Table).Rows[0].Cells[0].Borders, cell.Borders)
	assert.Equal(t, ir.Color{R: 230, G: 231, B: 232}, cell.Shading)
	assert.Equal(t, ir.Padding{Top: 4, Right: 5, Bottom: 6, Left: 7}, cell.Padding)
}

func TestSaveDOCXRejectsUnsupportedHighlightColor(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{
		ID:   "p1",
		Runs: []ir.Run{{Text: "Custom highlight", Attrs: ir.RunAttrs{Highlight: ir.Color{R: 1, G: 2, B: 3}}}},
	}}

	content, report, err := lean.SaveDOCX(document)
	assert.ErrorIs(t, err, lean.ErrUnsupportedDocument)
	assert.Nil(t, content)
	assert.Contains(t, report.Unsupported, "highlighting")
}

func TestSaveDOCXRoundTripsEmbeddedImage(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Image{
		ID:     "image1",
		Data:   []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a},
		Format: ir.ImagePNG,
		Width:  200,
		Height: 100,
		Alt:    "Quarterly revenue chart",
	}}

	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	assert.True(t, report.Editable)
	reopened, reopenedReport, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.True(t, reopenedReport.Editable)
	require.Len(t, reopened.Sections[0].Blocks, 1)
	image := reopened.Sections[0].Blocks[0].(*ir.Image)
	assert.Equal(t, document.Sections[0].Blocks[0].(*ir.Image).Data, image.Data)
	assert.Equal(t, ir.ImagePNG, image.Format)
	assert.Equal(t, 200.0, image.Width)
	assert.Equal(t, 100.0, image.Height)
	assert.Equal(t, "Quarterly revenue chart", image.Alt)
}

func TestOpenDOCXReadsExternalHyperlinks(t *testing.T) {
	content, err := os.ReadFile("testdata/fixtures/ooxml/python-docx/par-hyperlinks.docx")
	require.NoError(t, err)
	document, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)

	var found bool
	for _, block := range document.Sections[0].Blocks {
		paragraph, ok := block.(*ir.Paragraph)
		if !ok {
			continue
		}
		for _, run := range paragraph.Runs {
			if run.Text == "awesome hyperlink" && run.Hyperlink != nil {
				assert.Equal(t, "http://yahoo.com/", run.Hyperlink.URL)
				found = true
			}
		}
	}
	assert.True(t, found)
}

func TestSaveDOCXRoundTripsHyperlinks(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{ID: "p1", Runs: []ir.Run{
		{Text: "Website", Hyperlink: &ir.Hyperlink{URL: "https://whilesmart.com/docs?from=lean&format=docx"}},
		{Text: "Section", Hyperlink: &ir.Hyperlink{Bookmark: "overview"}},
	}}}

	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	assert.True(t, report.Editable)
	reopened, reopenedReport, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.True(t, reopenedReport.Editable)
	paragraph := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, paragraph.Runs, 2)
	require.NotNil(t, paragraph.Runs[0].Hyperlink)
	assert.Equal(t, "https://whilesmart.com/docs?from=lean&format=docx", paragraph.Runs[0].Hyperlink.URL)
	require.NotNil(t, paragraph.Runs[1].Hyperlink)
	assert.Equal(t, "overview", paragraph.Runs[1].Hyperlink.Bookmark)
}

func TestOpenDOCXReadsNumberingDefinitions(t *testing.T) {
	content, err := os.ReadFile("testdata/fixtures/ooxml/python-docx/num-having-numbering-part.docx")
	require.NoError(t, err)
	document, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.NotEmpty(t, document.Styles.Numbering)
	assert.Equal(t, "1", document.Styles.Numbering[0].ID)
	assert.Equal(t, ir.NumFormatBullet, document.Styles.Numbering[0].Levels[0].Format)
}

func TestSaveDOCXRoundTripsNumbering(t *testing.T) {
	document := ir.NewDocument()
	document.Styles.Numbering = []ir.NumberingDef{{
		ID: "42",
		Levels: []ir.NumberingLevel{
			{Level: 0, Format: ir.NumFormatDecimal, Text: "%1.", Start: 3},
			{Level: 1, Format: ir.NumFormatLowerAlpha, Text: "%2)", Start: 1},
		},
	}}
	document.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Numbering: &ir.NumberingRef{ID: "42", Level: 0}, Runs: []ir.Run{{Text: "Third item"}}},
		&ir.Paragraph{ID: "p2", Numbering: &ir.NumberingRef{ID: "42", Level: 1}, Runs: []ir.Run{{Text: "Nested item"}}},
	}

	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	assert.True(t, report.Editable)
	reopened, reopenedReport, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.True(t, reopenedReport.Editable)
	require.Len(t, reopened.Styles.Numbering, 1)
	assert.Equal(t, document.Styles.Numbering, reopened.Styles.Numbering)
	first := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	second := reopened.Sections[0].Blocks[1].(*ir.Paragraph)
	assert.Equal(t, &ir.NumberingRef{ID: "42", Level: 0}, first.Numbering)
	assert.Equal(t, &ir.NumberingRef{ID: "42", Level: 1}, second.Numbering)
}

func floatPointer(value float64) *float64 {
	return &value
}

func documentHasImage(document *ir.Document) bool {
	for _, section := range document.Sections {
		for _, block := range section.Blocks {
			if _, ok := block.(*ir.Image); ok {
				return true
			}
		}
	}
	return false
}
