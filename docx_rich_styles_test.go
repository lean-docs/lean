package lean_test

import (
	"encoding/json"
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestWordcraftHandbookStylesSurviveEditingAndDOCX(t *testing.T) {
	for _, fixture := range []string{"handbook.docx", "handbook-design.docx"} {
		t.Run(fixture, func(t *testing.T) {
			content, err := os.ReadFile("testdata/wordcraft/" + fixture)
			require.NoError(t, err)
			document, report, err := lean.OpenDOCX(content)
			require.NoError(t, err)
			require.True(t, report.Editable, report.Unsupported)
			style := document.Styles.Named["GridTable4AccentBlue"]
			require.NotNil(t, style.Table)
			require.Equal(t, ir.Color{R: 0x15, G: 0x60, B: 0x82}, style.Table.Conditions["firstRow"].Shading)
			require.True(t, style.Table.Conditions["firstCol"].RunAttrs.Bold)
			require.NotNil(t, document.Sections[0].Footer.Blocks[0].(*ir.Paragraph).Runs[0].Field)
			first := document.Sections[0].Blocks[0].(*ir.Paragraph)
			edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.replace", Block: first.ID, Start: 0, End: 0, Text: "Updated "}})
			require.NoError(t, err)
			encoded, err := json.Marshal(edited)
			require.NoError(t, err)
			var decoded ir.Document
			require.NoError(t, json.Unmarshal(encoded, &decoded))
			output, report, err := lean.SaveDOCX(&decoded)
			require.NoError(t, err, report.Unsupported)
			reopened, report, err := lean.OpenDOCX(output)
			require.NoError(t, err)
			require.True(t, report.Editable, report.Unsupported)
			require.Equal(t, document.Styles.Named, reopened.Styles.Named)
			require.Equal(t, decoded.Sections[0].Footer, reopened.Sections[0].Footer)
			require.Equal(t, decoded.Sections[0].Header, reopened.Sections[0].Header)
			require.Contains(t, reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs[0].Text, "Updated ")
			for index, block := range decoded.Sections[0].Blocks {
				require.Equal(t, block, reopened.Sections[0].Blocks[index])
			}
		})
	}
}

func TestImportedFooterFieldEditsKeepInstructions(t *testing.T) {
	content, err := os.ReadFile("testdata/wordcraft/handbook.docx")
	require.NoError(t, err)
	document, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	footer := document.Sections[0].Footer
	paragraph := footer.Blocks[0].(*ir.Paragraph)
	for _, offset := range []int{0, 1} {
		edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.replace", Part: "footer", Variant: footer.Type, Block: paragraph.ID, Start: offset, End: offset, Text: "Page "}})
		require.NoError(t, err)
		output, report, err := lean.SaveDOCX(edited)
		require.NoError(t, err, report.Unsupported)
		reopened, _, err := lean.OpenDOCX(output)
		require.NoError(t, err)
		runs := reopened.Sections[0].Footer.Blocks[0].(*ir.Paragraph).Runs
		var fields []ir.FieldElement
		for _, run := range runs {
			if run.Field != nil {
				fields = append(fields, *run.Field)
			}
		}
		require.Equal(t, []ir.FieldElement{{Type: "begin"}, {Type: "instruction", Instruction: " PAGE "}, {Type: "separate"}, {Type: "end"}}, fields)
		if offset == 0 {
			require.Equal(t, "Page ", runs[0].Text)
		} else {
			require.Equal(t, "Page ", runs[len(runs)-1].Text)
		}
	}
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.replace", Part: "footer", Variant: footer.Type, Block: paragraph.ID, Start: 0, End: 1}})
	require.NoError(t, err)
	for _, run := range edited.Sections[0].Footer.Blocks[0].(*ir.Paragraph).Runs {
		require.Nil(t, run.Field)
	}
}
