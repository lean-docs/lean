package ooxml

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
	"testing"
)

// buildDocx assembles a minimal valid .docx in memory whose word/document.xml
// body contains the given raw <w:p>...</w:p> XML. Used to create per-test
// fixtures without shipping binary blobs.
func buildDocx(t *testing.T, paragraphs string) []byte {
	t.Helper()
	doc := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>%s</w:body>
</w:document>`, paragraphs)

	contentTypes := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
  <Default Extension="xml" ContentType="application/xml"/>
  <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
  <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`

	rels := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
  <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	write := func(name, body string) {
		f, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip.Create(%s): %v", name, err)
		}
		if _, err := f.Write([]byte(body)); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	write("[Content_Types].xml", contentTypes)
	write("_rels/.rels", rels)
	write("word/document.xml", doc)
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

// para wraps run XML in a <w:p>.
func para(runs ...string) string {
	return "<w:p>" + strings.Join(runs, "") + "</w:p>"
}

// runXML builds a run with the given rPr fragment and text. rPr may be "".
func runXML(rPr, text string) string {
	props := ""
	if rPr != "" {
		props = "<w:rPr>" + rPr + "</w:rPr>"
	}
	return "<w:r>" + props + "<w:t>" + text + "</w:t></w:r>"
}
