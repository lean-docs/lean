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
	chain := []Style{style}
	seen := map[string]bool{name: true}
	for style.BasedOn != "" {
		if seen[style.BasedOn] {
			return Style{}, false
		}
		seen[style.BasedOn] = true
		parent, exists := ss.Named[style.BasedOn]
		if !exists {
			break
		}
		chain = append(chain, parent)
		style = parent
	}
	merged := chain[len(chain)-1]
	merged.RunAttrs = ss.Defaults
	merged.ParaAttrs = ParaAttrs{}
	for index := len(chain) - 1; index >= 0; index-- {
		current := chain[index]
		merged.RunAttrs = merged.RunAttrs.Overlay(current.RunAttrs)
		merged.ParaAttrs = merged.ParaAttrs.Overlay(current.ParaAttrs)
	}
	merged.ID = chain[0].ID
	merged.Name = chain[0].Name
	merged.BasedOn = chain[0].BasedOn
	merged.IsChar = chain[0].IsChar
	return merged, true
}
