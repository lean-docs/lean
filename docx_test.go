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

func TestOpenDOCXRejectsLossyImageEditing(t *testing.T) {
	content, err := os.ReadFile("testdata/fixtures/ooxml/python-docx/having-images.docx")
	require.NoError(t, err)
	_, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	assert.False(t, report.Editable)
	assert.Contains(t, report.Unsupported, "images")
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
