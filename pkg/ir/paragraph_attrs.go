package ir

import "encoding/json"

const (
	explicitAlignment uint8 = 1 << iota
	explicitSpacing
	explicitIndent
	explicitOutline
	explicitKeepTogether
	explicitKeepNext
	explicitPageBreak
	explicitBiDi
)

func (attrs *ParaAttrs) SetAlignment(value Alignment) {
	attrs.Align = value
	if value == AlignLeft {
		attrs.explicitProperties |= explicitAlignment
	} else {
		attrs.explicitProperties &^= explicitAlignment
	}
}
func (attrs ParaAttrs) AlignmentSpecified() bool {
	return attrs.Align != AlignLeft || attrs.explicitProperties&explicitAlignment != 0
}
func (attrs *ParaAttrs) SetSpacing(value Spacing) {
	attrs.Spacing = value
	if value == (Spacing{}) {
		attrs.explicitProperties |= explicitSpacing
	} else {
		attrs.explicitProperties &^= explicitSpacing
	}
}
func (attrs ParaAttrs) SpacingSpecified() bool {
	return attrs.Spacing != (Spacing{}) || attrs.explicitProperties&explicitSpacing != 0
}
func (attrs *ParaAttrs) SetIndent(value Indent) {
	attrs.Indent = value
	if value == (Indent{}) {
		attrs.explicitProperties |= explicitIndent
	} else {
		attrs.explicitProperties &^= explicitIndent
	}
}
func (attrs ParaAttrs) IndentSpecified() bool {
	return attrs.Indent != (Indent{}) || attrs.explicitProperties&explicitIndent != 0
}
func (attrs *ParaAttrs) SetOutlineLevel(value int) {
	attrs.OutlineLevel = value
	if value == 0 {
		attrs.explicitProperties |= explicitOutline
	} else {
		attrs.explicitProperties &^= explicitOutline
	}
}
func (attrs ParaAttrs) OutlineLevelSpecified() bool {
	return attrs.OutlineLevel != 0 || attrs.explicitProperties&explicitOutline != 0
}

type paraAttrsAlias ParaAttrs

func (attrs ParaAttrs) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		paraAttrsAlias
		Align           *Alignment `json:"align,omitempty"`
		Spacing         *Spacing   `json:"spacing,omitempty"`
		Indent          *Indent    `json:"indent,omitempty"`
		Outline         *int       `json:"outline_level,omitempty"`
		BiDi            *bool      `json:"bidi,omitempty"`
		PageBreakBefore *bool      `json:"page_break_before,omitempty"`
		KeepWithNext    *bool      `json:"keep_with_next,omitempty"`
		KeepTogether    *bool      `json:"keep_together,omitempty"`
	}{paraAttrsAlias(attrs), specifiedValue(attrs.Align, attrs.AlignmentSpecified()), specifiedValue(attrs.Spacing, attrs.SpacingSpecified()), specifiedValue(attrs.Indent, attrs.IndentSpecified()), specifiedValue(attrs.OutlineLevel, attrs.OutlineLevelSpecified()), specifiedValue(attrs.BiDi, attrs.BiDiSpecified()), specifiedValue(attrs.PageBreakBefore, attrs.PageBreakBeforeSpecified()), specifiedValue(attrs.KeepWithNext, attrs.KeepWithNextSpecified()), specifiedValue(attrs.KeepTogether, attrs.KeepTogetherSpecified())})
}

func (attrs *ParaAttrs) UnmarshalJSON(content []byte) error {
	decoded := paraAttrsAlias{}
	input := struct {
		*paraAttrsAlias
		Align           *Alignment `json:"align"`
		Spacing         *Spacing   `json:"spacing"`
		Indent          *Indent    `json:"indent"`
		Outline         *int       `json:"outline_level"`
		BiDi            *bool      `json:"bidi"`
		PageBreakBefore *bool      `json:"page_break_before"`
		KeepWithNext    *bool      `json:"keep_with_next"`
		KeepTogether    *bool      `json:"keep_together"`
	}{paraAttrsAlias: &decoded}
	if err := json.Unmarshal(content, &input); err != nil {
		return err
	}
	*attrs = ParaAttrs(decoded)
	if input.BiDi != nil {
		attrs.SetBiDi(*input.BiDi)
	}
	if input.PageBreakBefore != nil {
		attrs.SetPageBreakBefore(*input.PageBreakBefore)
	}
	if input.KeepWithNext != nil {
		attrs.SetKeepWithNext(*input.KeepWithNext)
	}
	if input.KeepTogether != nil {
		attrs.SetKeepTogether(*input.KeepTogether)
	}
	if input.Align != nil {
		attrs.SetAlignment(*input.Align)
	}
	if input.Spacing != nil {
		attrs.SetSpacing(*input.Spacing)
	}
	if input.Indent != nil {
		attrs.SetIndent(*input.Indent)
	}
	if input.Outline != nil {
		attrs.SetOutlineLevel(*input.Outline)
	}
	return nil
}

func (base ParaAttrs) Overlay(override ParaAttrs) ParaAttrs {
	if override.AlignmentSpecified() {
		base.SetAlignment(override.Align)
	}
	if override.SpacingSpecified() {
		base.SetSpacing(override.Spacing)
	}
	if override.IndentSpecified() {
		base.SetIndent(override.Indent)
	}
	if override.OutlineLevelSpecified() {
		base.SetOutlineLevel(override.OutlineLevel)
	}
	if override.TabStops != nil {
		base.TabStops = append([]TabStop{}, override.TabStops...)
	}
	if override.KeepTogetherSpecified() {
		base.SetKeepTogether(override.KeepTogether)
	}
	if override.KeepWithNextSpecified() {
		base.SetKeepWithNext(override.KeepWithNext)
	}
	if override.PageBreakBeforeSpecified() {
		base.SetPageBreakBefore(override.PageBreakBefore)
	}
	if override.Borders != (ParaBorders{}) {
		base.Borders = override.Borders
	}
	if override.Shading != (Color{}) {
		base.Shading = override.Shading
	}
	if override.BiDiSpecified() {
		base.SetBiDi(override.BiDi)
	}
	return base
}

func (attrs *ParaAttrs) SetKeepTogether(value bool) {
	attrs.KeepTogether = value
	if value {
		attrs.explicitProperties &^= explicitKeepTogether
	} else {
		attrs.explicitProperties |= explicitKeepTogether
	}
}
func (attrs ParaAttrs) KeepTogetherSpecified() bool {
	return attrs.KeepTogether || attrs.explicitProperties&explicitKeepTogether != 0
}

func (attrs *ParaAttrs) SetKeepWithNext(value bool) {
	attrs.KeepWithNext = value
	if value {
		attrs.explicitProperties &^= explicitKeepNext
	} else {
		attrs.explicitProperties |= explicitKeepNext
	}
}
func (attrs ParaAttrs) KeepWithNextSpecified() bool {
	return attrs.KeepWithNext || attrs.explicitProperties&explicitKeepNext != 0
}

func (attrs *ParaAttrs) SetPageBreakBefore(value bool) {
	attrs.PageBreakBefore = value
	if value {
		attrs.explicitProperties &^= explicitPageBreak
	} else {
		attrs.explicitProperties |= explicitPageBreak
	}
}
func (attrs ParaAttrs) PageBreakBeforeSpecified() bool {
	return attrs.PageBreakBefore || attrs.explicitProperties&explicitPageBreak != 0
}

func (attrs *ParaAttrs) SetBiDi(value bool) {
	attrs.BiDi = value
	if value {
		attrs.explicitProperties &^= explicitBiDi
	} else {
		attrs.explicitProperties |= explicitBiDi
	}
}
func (attrs ParaAttrs) BiDiSpecified() bool {
	return attrs.BiDi || attrs.explicitProperties&explicitBiDi != 0
}
