// Package ir defines the internal document representation.
//
// The IR is the source of truth for all document operations.
// All parsers produce it; all renderers and exporters consume it.
// It is fully serializable to/from JSON without loss.
package ir

import "time"

// BlockType identifies the kind of block element.
type BlockType int

const (
	BlockParagraph BlockType = iota
	BlockTable
	BlockImage
	BlockBookmark
)

// Orientation represents page orientation.
type Orientation int

const (
	Portrait Orientation = iota
	Landscape
)

// Alignment represents text or element alignment.
type Alignment int

const (
	AlignLeft Alignment = iota
	AlignCenter
	AlignRight
	AlignJustify
)

// VAlignment represents vertical alignment.
type VAlignment int

const (
	VAlignTop VAlignment = iota
	VAlignCenter
	VAlignBottom
)

// UnderlineStyle represents underline variants.
type UnderlineStyle int

const (
	UnderlineNone UnderlineStyle = iota
	UnderlineSingle
	UnderlineDouble
	UnderlineWavy
	UnderlineDash
)

// BaselineShift represents superscript/subscript.
type BaselineShift int

const (
	BaselineNone BaselineShift = iota
	BaselineSuperscript
	BaselineSubscript
)

// BreakType represents inline break types.
type BreakType int

const (
	BreakNone BreakType = iota
	BreakLine
	BreakPage
	BreakColumn
)

// LineRule represents line spacing rule.
type LineRule int

const (
	LineRuleAuto LineRule = iota
	LineRuleExact
	LineRuleAtLeast
)

// TabAlignment represents tab stop alignment.
type TabAlignment int

const (
	TabAlignLeft TabAlignment = iota
	TabAlignRight
	TabAlignCenter
	TabAlignDecimal
)

// TabLeader represents tab leader character.
type TabLeader int

const (
	TabLeaderNone TabLeader = iota
	TabLeaderDot
	TabLeaderDash
	TabLeaderUnderscore
)

// HeaderFooterKind represents header/footer types.
type HeaderFooterKind int

const (
	HeaderFooterDefault HeaderFooterKind = iota
	HeaderFooterFirst
	HeaderFooterEven
)

// NumberFormat represents numbering formats.
type NumberFormat int

const (
	NumFormatBullet NumberFormat = iota
	NumFormatDecimal
	NumFormatLowerAlpha
	NumFormatUpperAlpha
	NumFormatLowerRoman
	NumFormatUpperRoman
)

// ImageFormat represents image file formats.
type ImageFormat int

const (
	ImagePNG ImageFormat = iota
	ImageJPEG
	ImageGIF
	ImageSVG
)

// FloatMode represents image float behavior.
type FloatMode int

const (
	FloatNone FloatMode = iota
	FloatLeft
	FloatRight
)

// Document is the root of the IR.
type Document struct {
	Meta     DocumentMeta `json:"meta"`
	Styles   StyleSheet   `json:"styles"`
	Sections []Section    `json:"sections"`
}

// DocumentMeta contains document metadata.
type DocumentMeta struct {
	IRVersion  string    `json:"ir_version"`
	Title      string    `json:"title"`
	Author     string    `json:"author"`
	Language   string    `json:"language"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

// StyleSheet holds document styles.
type StyleSheet struct {
	Defaults  RunAttrs         `json:"defaults"`
	Named     map[string]Style `json:"named"`
	Theme     Theme            `json:"theme"`
	Numbering []NumberingDef   `json:"numbering"`
}

// Theme holds theme colors and fonts.
type Theme struct {
	Colors map[string]Color `json:"colors"`
	Fonts  ThemeFonts       `json:"fonts"`
}

// ThemeFonts holds theme font definitions.
type ThemeFonts struct {
	Major string `json:"major"`
	Minor string `json:"minor"`
}

// Style represents a named style.
type Style struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BasedOn   string    `json:"based_on,omitempty"`
	ParaAttrs ParaAttrs `json:"para_attrs"`
	RunAttrs  RunAttrs  `json:"run_attrs"`
	IsChar    bool      `json:"is_char"`
}

// Section represents a document section.
type Section struct {
	Properties PageProperties `json:"properties"`
	Columns    []Column       `json:"columns"`
	Header     *HeaderFooter  `json:"header,omitempty"`
	Footer     *HeaderFooter  `json:"footer,omitempty"`
	Blocks     []Block        `json:"blocks"`
}

// Column represents a column definition.
type Column struct {
	Width   float64 `json:"width"`
	Spacing float64 `json:"spacing"`
}

// PageProperties defines page dimensions and margins.
type PageProperties struct {
	Width        float64     `json:"width"`
	Height       float64     `json:"height"`
	MarginTop    float64     `json:"margin_top"`
	MarginBottom float64     `json:"margin_bottom"`
	MarginLeft   float64     `json:"margin_left"`
	MarginRight  float64     `json:"margin_right"`
	Orientation  Orientation `json:"orientation"`
}

// HeaderFooter represents header or footer content.
type HeaderFooter struct {
	Blocks []Block          `json:"blocks"`
	Kind   HeaderFooterKind `json:"kind"`
}

// Block is the interface for block-level elements.
type Block interface {
	BlockType() BlockType
	BlockID() string
}

// Paragraph is a block of text runs.
type Paragraph struct {
	ID        string        `json:"id"`
	Style     string        `json:"style,omitempty"`
	Runs      []Run         `json:"runs"`
	Para      ParaAttrs     `json:"para"`
	Numbering *NumberingRef `json:"numbering,omitempty"`
	Footnotes []Footnote    `json:"footnotes,omitempty"`
}

func (p *Paragraph) BlockType() BlockType { return BlockParagraph }
func (p *Paragraph) BlockID() string      { return p.ID }

// ParaAttrs holds paragraph formatting attributes.
type ParaAttrs struct {
	Align           Alignment   `json:"align"`
	Spacing         Spacing     `json:"spacing"`
	Indent          Indent      `json:"indent"`
	TabStops        []TabStop   `json:"tab_stops,omitempty"`
	KeepTogether    bool        `json:"keep_together,omitempty"`
	KeepWithNext    bool        `json:"keep_with_next,omitempty"`
	PageBreakBefore bool        `json:"page_break_before,omitempty"`
	OutlineLevel    int         `json:"outline_level,omitempty"`
	Borders         ParaBorders `json:"borders"`
	Shading         Color       `json:"shading"`
	BiDi            bool        `json:"bidi,omitempty"`
}

// ParaBorders holds paragraph border definitions.
type ParaBorders struct {
	Top    Border `json:"top"`
	Bottom Border `json:"bottom"`
	Left   Border `json:"left"`
	Right  Border `json:"right"`
}

// Border represents a single border edge.
type Border struct {
	Style string  `json:"style,omitempty"`
	Width float64 `json:"width,omitempty"`
	Color Color   `json:"color"`
}

// Run is an inline text segment with formatting.
type Run struct {
	Text      string     `json:"text"`
	Attrs     RunAttrs   `json:"attrs"`
	Hyperlink *Hyperlink `json:"hyperlink,omitempty"`
	Break     BreakType  `json:"break,omitempty"`
}

// RunAttrs holds character formatting attributes.
type RunAttrs struct {
	Bold      bool           `json:"bold,omitempty"`
	Italic    bool           `json:"italic,omitempty"`
	Underline UnderlineStyle `json:"underline,omitempty"`
	Strike    bool           `json:"strike,omitempty"`
	SmallCaps bool           `json:"small_caps,omitempty"`
	AllCaps   bool           `json:"all_caps,omitempty"`
	FontSize  float64        `json:"font_size,omitempty"`
	FontName  string         `json:"font_name,omitempty"`
	Color     Color          `json:"color"`
	Highlight Color          `json:"highlight"`
	Baseline  BaselineShift  `json:"baseline,omitempty"`
	Tracking  float64        `json:"tracking,omitempty"`
	Language  string         `json:"language,omitempty"`
}

// Color represents an RGB color.
type Color struct {
	R    uint8 `json:"r"`
	G    uint8 `json:"g"`
	B    uint8 `json:"b"`
	None bool  `json:"none,omitempty"`
}

// Hyperlink represents a link target.
type Hyperlink struct {
	URL      string `json:"url,omitempty"`
	Bookmark string `json:"bookmark,omitempty"`
}

// TabStop represents a tab stop definition.
type TabStop struct {
	Position  float64      `json:"position"`
	Alignment TabAlignment `json:"alignment"`
	Leader    TabLeader    `json:"leader"`
}

// Spacing represents paragraph spacing.
type Spacing struct {
	Before   float64  `json:"before"`
	After    float64  `json:"after"`
	Line     float64  `json:"line"`
	LineRule LineRule  `json:"line_rule"`
}

// Indent represents paragraph indentation.
type Indent struct {
	Left      float64 `json:"left"`
	Right     float64 `json:"right"`
	FirstLine float64 `json:"first_line"`
	Hanging   float64 `json:"hanging"`
}

// NumberingRef references a numbering definition from a paragraph.
type NumberingRef struct {
	ID      string `json:"id"`
	Level   int    `json:"level"`
	Restart *int   `json:"restart,omitempty"`
}

// NumberingDef defines a numbering scheme.
type NumberingDef struct {
	ID     string           `json:"id"`
	Levels []NumberingLevel `json:"levels"`
}

// NumberingLevel defines a single level of a numbering scheme.
type NumberingLevel struct {
	Level    int          `json:"level"`
	Format   NumberFormat `json:"format"`
	Text     string       `json:"text"`
	Align    Alignment    `json:"align"`
	Indent   Indent       `json:"indent"`
	RunAttrs RunAttrs     `json:"run_attrs"`
	Start    int          `json:"start"`
}

// Table is a block-level table element.
type Table struct {
	ID           string       `json:"id"`
	Rows         []TableRow   `json:"rows"`
	ColumnWidths []float64    `json:"column_widths"`
	Align        Alignment    `json:"align"`
	Borders      TableBorders `json:"borders"`
	Shading      Color        `json:"shading"`
	Style        string       `json:"style,omitempty"`
}

func (t *Table) BlockType() BlockType { return BlockTable }
func (t *Table) BlockID() string      { return t.ID }

// TableBorders holds table-level borders.
type TableBorders struct {
	Top      Border `json:"top"`
	Bottom   Border `json:"bottom"`
	Left     Border `json:"left"`
	Right    Border `json:"right"`
	InsideH  Border `json:"inside_h"`
	InsideV  Border `json:"inside_v"`
}

// TableRow represents a row in a table.
type TableRow struct {
	Cells    []TableCell `json:"cells"`
	IsHeader bool        `json:"is_header,omitempty"`
	Height   *float64    `json:"height,omitempty"`
}

// TableCell represents a cell in a table row.
type TableCell struct {
	Blocks  []Block     `json:"blocks"`
	Borders CellBorders `json:"borders"`
	Shading Color       `json:"shading"`
	ColSpan int         `json:"col_span"`
	RowSpan int         `json:"row_span"`
	VAlign  VAlignment  `json:"v_align"`
	Width   float64     `json:"width"`
	Padding Padding     `json:"padding"`
}

// CellBorders holds cell-level borders.
type CellBorders struct {
	Top    Border `json:"top"`
	Bottom Border `json:"bottom"`
	Left   Border `json:"left"`
	Right  Border `json:"right"`
}

// Padding represents padding values.
type Padding struct {
	Top    float64 `json:"top"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
	Right  float64 `json:"right"`
}

// Image is a block-level image element.
type Image struct {
	ID     string      `json:"id"`
	Data   []byte      `json:"data"`
	Format ImageFormat `json:"format"`
	Width  float64     `json:"width"`
	Height float64     `json:"height"`
	Float  FloatMode   `json:"float"`
	Alt    string      `json:"alt,omitempty"`
}

func (i *Image) BlockType() BlockType { return BlockImage }
func (i *Image) BlockID() string      { return i.ID }

// Footnote represents a footnote.
type Footnote struct {
	ID     string  `json:"id"`
	Blocks []Block `json:"blocks"`
}

// Bookmark represents a document bookmark.
type Bookmark struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (b *Bookmark) BlockType() BlockType { return BlockBookmark }
func (b *Bookmark) BlockID() string      { return b.ID }
