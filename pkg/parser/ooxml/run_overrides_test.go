package ooxml

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"github.com/lean-docs/lean/pkg/ir"
	"io"
	"testing"
)

func TestCapturedLibreOfficeRunPropertiesPreserveExplicitOff(t *testing.T) {
	archive, err := zip.OpenReader("../../../testdata/fixtures/ooxml/libreoffice/hyperlink.docx")
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	for _, part := range archive.File {
		if part.Name != "word/document.xml" {
			continue
		}
		reader, err := part.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
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
			if !ok || start.Name.Local != "rPr" {
				continue
			}
			var props xmlRunProps
			if err := decoder.DecodeElement(&props, &start); err != nil {
				t.Fatal(err)
			}
			if props.Bold == nil || props.Bold.Val != "0" {
				continue
			}
			attrs := ir.RunAttrs{Bold: true}
			applyRunProps(&attrs, &props)
			if attrs.Bold || !attrs.BoldSpecified() {
				t.Fatal("captured explicit off value was lost")
			}
			content, err := json.Marshal(attrs)
			if err != nil {
				t.Fatal(err)
			}
			var serialized map[string]any
			json.Unmarshal(content, &serialized)
			if value, present := serialized["bold"]; !present || value != false {
				t.Fatalf("explicit off missing from source: %s", content)
			}
			return
		}
	}
	t.Fatal("captured LibreOffice off property was not found")
}
