package lean_test

import (
	"os"
	"strings"
	"testing"

	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
)

func TestReferenceNestedTextBoxSurvivesEditingAndDOCX(t *testing.T) {
	input, err := os.ReadFile("testdata/reference-layouts/nested-textbox.docx")
	require.NoError(t, err)
	document, report, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	require.True(t, report.Editable, report.Unsupported)
	box := document.Sections[0].Blocks[1].(*ir.Shape)
	require.Equal(t, "rect", box.Geometry)
	require.Equal(t, 180.0, box.Width)
	require.Equal(t, 80.0, box.Height)
	require.Equal(t, "page", box.Anchor.Horizontal.Relative)
	require.Equal(t, 72.0, box.Anchor.Horizontal.Offset)
	require.Equal(t, 100.0, box.Anchor.Vertical.Offset)
	require.Equal(t, &ir.Color{R: 240, G: 248, B: 240}, box.Fill)
	picture := box.Blocks[1].(*ir.Image)
	require.Equal(t, 40.0, picture.Width)
	require.Equal(t, 10.0, picture.Height)
	paragraph := box.Blocks[0].(*ir.Paragraph)
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.replace", Block: paragraph.ID, Start: 0, End: 6, Text: "Updated parent"}})
	require.NoError(t, err)
	require.Equal(t, "Parent", paragraph.Runs[1].Text)
	content, report, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.True(t, report.Editable, report.Unsupported)
	reopened, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.True(t, report.Editable, report.Unsupported)
	after := reopened.Sections[0].Blocks[1].(*ir.Shape)
	require.Equal(t, box.Width, after.Width)
	require.Equal(t, box.Height, after.Height)
	require.Equal(t, box.Anchor, after.Anchor)
	require.Equal(t, box.Fill, after.Fill)
	require.Equal(t, box.Stroke, after.Stroke)
	require.Equal(t, box.AutoFit, after.AutoFit)
	nested := after.Blocks[1].(*ir.Image)
	require.Equal(t, picture.Data, nested.Data)
	require.Equal(t, picture.Anchor, nested.Anchor)
	var text strings.Builder
	for _, run := range after.Blocks[0].(*ir.Paragraph).Runs {
		text.WriteString(run.Text)
	}
	require.Equal(t, "Updated parent", text.String())
	xml := editingPart(t, content, "word/document.xml")
	require.Contains(t, xml, `<wps:txbx>`)
	require.Contains(t, xml, `<wp:anchor`)
	require.Contains(t, xml, `<wp:inline`)
	require.Contains(t, xml, `<a:srgbClr val="F0F8F0"/>`)
}

func TestDrawingCommandsPreserveReferenceMediaAndAreAtomic(t *testing.T) {
	input, err := os.ReadFile("testdata/reference-layouts/nested-textbox.docx")
	require.NoError(t, err)
	document, _, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	shape := document.Sections[0].Blocks[1].(*ir.Shape)
	source := shape.Blocks[1].(*ir.Image)
	picture := *source
	picture.ID = "added-picture"
	anchor := *source.Anchor
	anchor.ParagraphID = document.Sections[0].Blocks[0].BlockID()
	anchor.Offset = 0
	anchor.Floating = true
	anchor.Horizontal = ir.DrawingPosition{Relative: "page", Offset: 144}
	anchor.Vertical = ir.DrawingPosition{Relative: "page", Offset: 240}
	anchor.Wrap = "square"
	anchor.Crop = ir.Padding{Left: 0.1, Right: 0.1}
	anchor.Rotation = 15
	anchor.FlipHorizontal = true
	picture.Anchor = &anchor
	width := 200.0
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{
		{ID: "image.insert", At: 2, Image: &picture},
		{ID: "drawing.layout", Block: shape.ID, Drawing: &lean.DrawingLayout{Width: &width}},
	})
	require.NoError(t, err)
	require.Equal(t, 180.0, shape.Width)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	var restoredShape *ir.Shape
	for _, block := range reopened.Sections[0].Blocks {
		if shape, ok := block.(*ir.Shape); ok {
			restoredShape = shape
		}
	}
	require.NotNil(t, restoredShape)
	require.Equal(t, 200.0, restoredShape.Width)
	var inserted *ir.Image
	for _, block := range reopened.Sections[0].Blocks {
		if image, ok := block.(*ir.Image); ok {
			inserted = image
		}
	}
	require.NotNil(t, inserted)
	require.Equal(t, source.Data, inserted.Data)
	require.Equal(t, anchor.Horizontal, inserted.Anchor.Horizontal)
	require.Equal(t, anchor.Crop, inserted.Anchor.Crop)
	require.Equal(t, 15.0, inserted.Anchor.Rotation)
	require.True(t, inserted.Anchor.FlipHorizontal)
	xml := editingPart(t, content, "word/document.xml")
	for _, id := range []string{`<wp:docPr id="1"`, `<wp:docPr id="2"`, `<wp:docPr id="3"`} {
		require.Equal(t, 1, strings.Count(xml, id))
	}
	zero := 0.0
	_, err = lean.ApplyDOCXCommands(edited, []lean.DocumentCommand{{ID: "drawing.layout", Block: shape.ID, Drawing: &lean.DrawingLayout{Width: &zero}}})
	require.Error(t, err)
	require.Equal(t, 200.0, edited.Sections[0].Blocks[1].(*ir.Shape).Width)
	removed, err := lean.ApplyDOCXCommands(edited, []lean.DocumentCommand{{ID: "drawing.delete", Block: "added-picture"}})
	require.NoError(t, err)
	require.Len(t, removed.Sections[0].Blocks, 2)
}

func TestHeaderOptionsCanBeDisabledWithoutDeletingVariantContent(t *testing.T) {
	document := lean.NewDOCX("Body")
	document.Sections[0].Headers = []*ir.HeaderFooter{{Type: ir.HeaderFooterFirst, Blocks: []ir.Block{editingParagraph("first", "First page header")}}, {Type: ir.HeaderFooterEven, Blocks: []ir.Block{editingParagraph("even", "Even page header")}}}
	disabled := false
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "header.options", EvenAndOddHeaders: &disabled, DifferentFirstPage: &disabled}})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.NotContains(t, editingPart(t, content, "word/document.xml"), "<w:titlePg")
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.NotNil(t, reopened.EvenAndOddHeaders)
	require.False(t, *reopened.EvenAndOddHeaders)
	require.NotNil(t, reopened.Sections[0].Properties.DifferentFirstPage)
	require.False(t, *reopened.Sections[0].Properties.DifferentFirstPage)
	require.NotNil(t, reopened.Sections[0].Header)
	require.Len(t, reopened.Sections[0].Headers, 1)
	enabled := true
	edited, err = lean.ApplyDOCXCommands(reopened, []lean.DocumentCommand{{ID: "header.options", EvenAndOddHeaders: &enabled, DifferentFirstPage: &enabled}})
	require.NoError(t, err)
	content, _, err = lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.Contains(t, editingPart(t, content, "word/settings.xml"), "<w:evenAndOddHeaders/>")
	require.Contains(t, editingPart(t, content, "word/document.xml"), "<w:titlePg/>")
}

func TestThreeColumnReferenceHeaderFooterEditing(t *testing.T) {
	input, err := os.ReadFile("testdata/reference-layouts/three-column-header-footer.docx")
	require.NoError(t, err)
	document, report, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	require.True(t, report.Editable, report.Unsupported)
	header := document.Sections[0].Header.Blocks[0].(*ir.Table)
	footer := document.Sections[0].Footer.Blocks[0].(*ir.Table)
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{
		{ID: "text.replace", Part: "header", Block: header.Rows[0].Cells[0].Blocks[0].BlockID(), Text: "Header edit "},
		{ID: "text.replace", Part: "footer", Block: footer.Rows[0].Cells[0].Blocks[0].BlockID(), Text: "Footer edit "},
	})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	reopened, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.True(t, report.Editable, report.Unsupported)
	for index, part := range []*ir.HeaderFooter{reopened.Sections[0].Header, reopened.Sections[0].Footer} {
		table := part.Blocks[0].(*ir.Table)
		require.Equal(t, []float64{144, 144, 144}, table.ColumnWidths)
		for cellIndex, cell := range table.Rows[0].Cells {
			paragraph := cell.Blocks[0].(*ir.Paragraph)
			require.Equal(t, ir.Alignment(cellIndex), paragraph.Para.Align)
			if cellIndex == 1 {
				require.True(t, paragraph.Runs[0].Attrs.Bold)
			}
		}
		text := ""
		for _, run := range table.Rows[0].Cells[0].Blocks[0].(*ir.Paragraph).Runs {
			text += run.Text
		}
		require.Equal(t, []string{"Header edit Left text", "Footer edit Left text"}[index], text)
	}
	require.Equal(t, "Body text", reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs[0].Text)
}
