package lean_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"encoding/xml"
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"io"
	"reflect"
	"testing"
)

func TestDOCXPreservesEveryStandardUnderlineStyle(t *testing.T) {
	expected := []string{"none", "single", "double", "wave", "dash", "words", "thick", "dotted", "dottedHeavy", "dashedHeavy", "dashLong", "dashLongHeavy", "dotDash", "dashDotHeavy", "dotDotDash", "dashDotDotHeavy", "wavyHeavy", "wavyDouble"}
	document := ir.NewDocument()
	paragraph := &ir.Paragraph{ID: "underlines"}
	for index, name := range expected {
		var attrs ir.RunAttrs
		attrs.SetUnderline(ir.UnderlineStyle(index))
		paragraph.Runs = append(paragraph.Runs, ir.Run{Text: name + " ", Attrs: attrs})
	}
	document.Sections[0].Blocks = []ir.Block{paragraph}
	content, report, err := lean.SaveDOCX(document)
	if err != nil || !report.Editable {
		t.Fatalf("save: %v %+v", err, report)
	}
	archive, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatal(err)
	}
	var names []string
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
			element, ok := token.(xml.StartElement)
			if !ok || element.Name.Local != "u" {
				continue
			}
			for _, attribute := range element.Attr {
				if attribute.Name.Local == "val" {
					names = append(names, attribute.Value)
				}
			}
		}
		reader.Close()
	}
	if !reflect.DeepEqual(names, expected) {
		t.Fatalf("underline values: %v", names)
	}
	reopened, report, err := lean.OpenDOCX(content)
	if err != nil || !report.Editable {
		t.Fatalf("open: %v %+v", err, report)
	}
	runs := reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs
	if len(runs) != len(expected) {
		t.Fatalf("runs: %d", len(runs))
	}
	for index, run := range runs {
		if run.Attrs.Underline != ir.UnderlineStyle(index) || !run.Attrs.UnderlineSpecified() {
			t.Fatalf("lost %s: %+v", expected[index], run.Attrs)
		}
	}
}

func TestDOCXPreservesExplicitUnderlineBaselineAndSpacingResets(t *testing.T) {
	var attrs ir.RunAttrs
	attrs.SetUnderline(ir.UnderlineNone)
	attrs.SetBaseline(ir.BaselineNone)
	attrs.SetTracking(0)
	content, err := json.Marshal(attrs)
	if err != nil {
		t.Fatal(err)
	}
	var restored ir.RunAttrs
	if err := json.Unmarshal(content, &restored); err != nil {
		t.Fatal(err)
	}
	base := ir.RunAttrs{Underline: ir.UnderlineDouble, Baseline: ir.BaselineSuperscript, Tracking: 2}
	resolved := base.Overlay(restored)
	if resolved.Underline != ir.UnderlineNone || resolved.Baseline != ir.BaselineNone || resolved.Tracking != 0 || !resolved.UnderlineSpecified() || !resolved.BaselineSpecified() || !resolved.TrackingSpecified() {
		t.Fatalf("lost resets: %+v", resolved)
	}
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{ID: "reset", Runs: []ir.Run{{Text: "Plain", Attrs: restored}}}}
	binary, report, err := lean.SaveDOCX(document)
	if err != nil || !report.Editable {
		t.Fatalf("save: %v %+v", err, report)
	}
	reopened, _, err := lean.OpenDOCX(binary)
	if err != nil {
		t.Fatal(err)
	}
	actual := reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs[0].Attrs
	if actual.Underline != ir.UnderlineNone || actual.Baseline != ir.BaselineNone || actual.Tracking != 0 || !actual.UnderlineSpecified() || !actual.BaselineSpecified() || !actual.TrackingSpecified() {
		t.Fatalf("reopen lost resets: %+v", actual)
	}
}

func TestDOCXRejectsUnknownUnderlineStyle(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{ID: "invalid", Runs: []ir.Run{{Text: "Invalid", Attrs: ir.RunAttrs{Underline: ir.UnderlineStyle(999)}}}}}
	content, report, err := lean.SaveDOCX(document)
	if err == nil || report.Editable || content != nil {
		t.Fatal("unknown underline style was silently replaced")
	}
}
