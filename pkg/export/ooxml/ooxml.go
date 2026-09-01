// Package ooxml exports an IR document to .docx (OOXML) format.
package ooxml

import "github.com/lean-docs/lean/pkg/ir"

// Export exports an IR Document to .docx bytes.
func Export(doc *ir.Document) ([]byte, error) {
	return buildPackage(doc)
}
