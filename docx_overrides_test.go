package lean_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"io"
	"testing"
)

func TestExplicitOffSurvivesSourceAndDOCXExport(t *testing.T) {
	document := ir.NewDocument()
	attrs := ir.RunAttrs{}
	attrs.SetBold(false)
	attrs.SetItalic(false)
	attrs.SetStrike(false)
	attrs.SetSmallCaps(false)
	attrs.SetAllCaps(false)
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{ID: "plain", Runs: []ir.Run{{Text: "Explicitly plain", Attrs: attrs}}}}
	source, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	var restored ir.Document
	if err := json.Unmarshal(source, &restored); err != nil {
		t.Fatal(err)
	}
	content, report, err := lean.SaveDOCX(&restored)
	if err != nil || !report.Editable {
		t.Fatalf("save: %v %+v", err, report)
	}
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, part := range archive.File {
		if part.Name != "word/document.xml" {
			continue
		}
		reader, err := part.Open()
		if err != nil {
			t.Fatal(err)
		}
		decoder := xml.NewDecoder(reader)
		for {
			token, err := decoder.Token()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatal(err)
			}
			start, ok := token.(xml.StartElement)
			if !ok {
				continue
			}
			for _, attribute := range start.Attr {
				if attribute.Name.Local == "val" && attribute.Value == "0" {
					found[start.Name.Local] = true
				}
			}
		}
		reader.Close()
	}
	for _, tag := range []string{"b", "i", "strike", "smallCaps", "caps"} {
		if !found[tag] {
			t.Fatalf("export lost explicit %s off", tag)
		}
	}
	reopened, _, err := lean.OpenDOCX(content)
	if err != nil {
		t.Fatal(err)
	}
	run := reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs[0]
	if run.Attrs.Bold || run.Attrs.Italic || run.Attrs.Strike || run.Attrs.SmallCaps || run.Attrs.AllCaps || !run.Attrs.BoldSpecified() || !run.Attrs.ItalicSpecified() || !run.Attrs.StrikeSpecified() || !run.Attrs.SmallCapsSpecified() || !run.Attrs.AllCapsSpecified() {
		t.Fatalf("reopen lost explicit off: %+v", run.Attrs)
	}
}
