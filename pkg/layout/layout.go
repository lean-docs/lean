// Package layout computes page layout from an IR document.
package layout

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// PageProps defines the page properties for layout.
type PageProps = ir.PageProperties

// Page represents a laid-out page.
type Page struct {
	Number int
	Width  float64
	Height float64
	Lines  []Line
}

// Line represents a laid-out line of content.
type Line struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Layout computes the page layout for a document.
func Layout(doc *ir.Document, props PageProps) ([]Page, error) {
	return nil, errors.New("layout: not implemented")
}
