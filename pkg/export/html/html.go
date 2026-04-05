// Package html exports an IR document to HTML.
package html

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// ErrNotImplemented is returned when the exporter is not yet implemented.
var ErrNotImplemented = errors.New("html exporter: not implemented")

// Export exports an IR Document to HTML bytes.
func Export(doc *ir.Document) ([]byte, error) {
	return nil, ErrNotImplemented
}
