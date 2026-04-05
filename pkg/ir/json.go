package ir

import (
	"encoding/json"
	"fmt"
)

// blockEnvelope wraps a Block for JSON serialization with type discrimination.
type blockEnvelope struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
}

// MarshalBlockSlice serializes a slice of Blocks to JSON.
func MarshalBlockSlice(blocks []Block) ([]byte, error) {
	envelopes := make([]blockEnvelope, len(blocks))
	for i, b := range blocks {
		typeName, err := blockTypeName(b)
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("marshal block %d: %w", i, err)
		}
		envelopes[i] = blockEnvelope{Type: typeName, Data: data}
	}
	return json.Marshal(envelopes)
}

// UnmarshalBlockSlice deserializes a slice of Blocks from JSON.
func UnmarshalBlockSlice(data []byte) ([]Block, error) {
	var envelopes []blockEnvelope
	if err := json.Unmarshal(data, &envelopes); err != nil {
		return nil, err
	}
	blocks := make([]Block, len(envelopes))
	for i, env := range envelopes {
		b, err := unmarshalBlock(env)
		if err != nil {
			return nil, fmt.Errorf("unmarshal block %d: %w", i, err)
		}
		blocks[i] = b
	}
	return blocks, nil
}

func blockTypeName(b Block) (string, error) {
	switch b.(type) {
	case *Paragraph:
		return "paragraph", nil
	case *Table:
		return "table", nil
	case *Image:
		return "image", nil
	case *Bookmark:
		return "bookmark", nil
	default:
		return "", fmt.Errorf("unknown block type: %T", b)
	}
}

func unmarshalBlock(env blockEnvelope) (Block, error) {
	switch env.Type {
	case "paragraph":
		var p Paragraph
		if err := json.Unmarshal(env.Data, &p); err != nil {
			return nil, err
		}
		return &p, nil
	case "table":
		var t Table
		if err := json.Unmarshal(env.Data, &t); err != nil {
			return nil, err
		}
		return &t, nil
	case "image":
		var img Image
		if err := json.Unmarshal(env.Data, &img); err != nil {
			return nil, err
		}
		return &img, nil
	case "bookmark":
		var b Bookmark
		if err := json.Unmarshal(env.Data, &b); err != nil {
			return nil, err
		}
		return &b, nil
	default:
		return nil, fmt.Errorf("unknown block type: %q", env.Type)
	}
}

// sectionJSON is the JSON representation of a Section with typed blocks.
type sectionJSON struct {
	Properties PageProperties  `json:"properties"`
	Columns    []Column        `json:"columns"`
	Header     *headerFooterJSON `json:"header,omitempty"`
	Footer     *headerFooterJSON `json:"footer,omitempty"`
	Blocks     json.RawMessage `json:"blocks"`
}

type headerFooterJSON struct {
	Blocks json.RawMessage  `json:"blocks"`
	Kind   HeaderFooterKind `json:"kind"`
}

// MarshalJSON implements custom JSON marshaling for Section.
func (s Section) MarshalJSON() ([]byte, error) {
	blocksData, err := MarshalBlockSlice(s.Blocks)
	if err != nil {
		return nil, err
	}

	sj := sectionJSON{
		Properties: s.Properties,
		Columns:    s.Columns,
		Blocks:     blocksData,
	}

	if s.Header != nil {
		hBlocks, err := MarshalBlockSlice(s.Header.Blocks)
		if err != nil {
			return nil, err
		}
		sj.Header = &headerFooterJSON{Blocks: hBlocks, Kind: s.Header.Kind}
	}

	if s.Footer != nil {
		fBlocks, err := MarshalBlockSlice(s.Footer.Blocks)
		if err != nil {
			return nil, err
		}
		sj.Footer = &headerFooterJSON{Blocks: fBlocks, Kind: s.Footer.Kind}
	}

	return json.Marshal(sj)
}

// UnmarshalJSON implements custom JSON unmarshaling for Section.
func (s *Section) UnmarshalJSON(data []byte) error {
	var sj sectionJSON
	if err := json.Unmarshal(data, &sj); err != nil {
		return err
	}

	s.Properties = sj.Properties
	s.Columns = sj.Columns

	if sj.Blocks != nil {
		blocks, err := UnmarshalBlockSlice(sj.Blocks)
		if err != nil {
			return err
		}
		s.Blocks = blocks
	}

	if sj.Header != nil {
		hBlocks, err := UnmarshalBlockSlice(sj.Header.Blocks)
		if err != nil {
			return err
		}
		s.Header = &HeaderFooter{Blocks: hBlocks, Kind: sj.Header.Kind}
	}

	if sj.Footer != nil {
		fBlocks, err := UnmarshalBlockSlice(sj.Footer.Blocks)
		if err != nil {
			return err
		}
		s.Footer = &HeaderFooter{Blocks: fBlocks, Kind: sj.Footer.Kind}
	}

	return nil
}

// tableCellJSON is the JSON representation of a TableCell with typed blocks.
type tableCellJSON struct {
	Blocks  json.RawMessage `json:"blocks"`
	Borders CellBorders     `json:"borders"`
	Shading Color           `json:"shading"`
	ColSpan int             `json:"col_span"`
	RowSpan int             `json:"row_span"`
	VAlign  VAlignment      `json:"v_align"`
	Width   float64         `json:"width"`
	Padding Padding         `json:"padding"`
}

// MarshalJSON implements custom JSON marshaling for TableCell.
func (c TableCell) MarshalJSON() ([]byte, error) {
	blocksData, err := MarshalBlockSlice(c.Blocks)
	if err != nil {
		return nil, err
	}
	return json.Marshal(tableCellJSON{
		Blocks:  blocksData,
		Borders: c.Borders,
		Shading: c.Shading,
		ColSpan: c.ColSpan,
		RowSpan: c.RowSpan,
		VAlign:  c.VAlign,
		Width:   c.Width,
		Padding: c.Padding,
	})
}

// UnmarshalJSON implements custom JSON unmarshaling for TableCell.
func (c *TableCell) UnmarshalJSON(data []byte) error {
	var cj tableCellJSON
	if err := json.Unmarshal(data, &cj); err != nil {
		return err
	}
	c.Borders = cj.Borders
	c.Shading = cj.Shading
	c.ColSpan = cj.ColSpan
	c.RowSpan = cj.RowSpan
	c.VAlign = cj.VAlign
	c.Width = cj.Width
	c.Padding = cj.Padding

	if cj.Blocks != nil {
		blocks, err := UnmarshalBlockSlice(cj.Blocks)
		if err != nil {
			return err
		}
		c.Blocks = blocks
	}
	return nil
}

// footnoteJSON is the JSON representation of a Footnote with typed blocks.
type footnoteJSON struct {
	ID     string          `json:"id"`
	Blocks json.RawMessage `json:"blocks"`
}

// MarshalJSON implements custom JSON marshaling for Footnote.
func (f Footnote) MarshalJSON() ([]byte, error) {
	blocksData, err := MarshalBlockSlice(f.Blocks)
	if err != nil {
		return nil, err
	}
	return json.Marshal(footnoteJSON{ID: f.ID, Blocks: blocksData})
}

// UnmarshalJSON implements custom JSON unmarshaling for Footnote.
func (f *Footnote) UnmarshalJSON(data []byte) error {
	var fj footnoteJSON
	if err := json.Unmarshal(data, &fj); err != nil {
		return err
	}
	f.ID = fj.ID
	if fj.Blocks != nil {
		blocks, err := UnmarshalBlockSlice(fj.Blocks)
		if err != nil {
			return err
		}
		f.Blocks = blocks
	}
	return nil
}
