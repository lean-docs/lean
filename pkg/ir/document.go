package ir

import "time"

const IRVersion = "1.0"

// NewDocument creates a valid zero-value document.
func NewDocument() *Document {
	now := time.Now().UTC()
	return &Document{
		Meta: DocumentMeta{
			IRVersion:  IRVersion,
			CreatedAt:  now,
			ModifiedAt: now,
		},
		Styles: StyleSheet{
			Named: make(map[string]Style),
			Theme: Theme{
				Colors: make(map[string]Color),
			},
		},
		Sections: []Section{
			{
				Properties: PageProperties{
					Width:        612, // US Letter in points
					Height:       792,
					MarginTop:    72,
					MarginBottom: 72,
					MarginLeft:   72,
					MarginRight:  72,
				},
			},
		},
	}
}

// ResolveStyle resolves a style by following the BasedOn chain.
func (ss *StyleSheet) ResolveStyle(name string) (Style, bool) {
	style, ok := ss.Named[name]
	if !ok {
		return Style{}, false
	}

	if style.BasedOn == "" {
		return style, true
	}

	parent, ok := ss.ResolveStyle(style.BasedOn)
	if !ok {
		return style, true
	}

	// Merge parent into style (style overrides parent)
	merged := parent
	merged.ID = style.ID
	merged.Name = style.Name
	merged.BasedOn = style.BasedOn
	merged.IsChar = style.IsChar

	// Only override non-zero values from child
	if style.RunAttrs.Bold {
		merged.RunAttrs.Bold = true
	}
	if style.RunAttrs.Italic {
		merged.RunAttrs.Italic = true
	}
	if style.RunAttrs.FontSize > 0 {
		merged.RunAttrs.FontSize = style.RunAttrs.FontSize
	}
	if style.RunAttrs.FontName != "" {
		merged.RunAttrs.FontName = style.RunAttrs.FontName
	}
	if style.ParaAttrs.Align != AlignLeft {
		merged.ParaAttrs.Align = style.ParaAttrs.Align
	}

	return merged, true
}
