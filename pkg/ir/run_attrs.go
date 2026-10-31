package ir

import "encoding/json"

const (
	offBold uint16 = 1 << iota
	offItalic
	offStrike
	offSmallCaps
	offAllCaps
	offUnderline
	offBaseline
	offTracking
	offColor
)

func (attrs *RunAttrs) SetBold(value bool) {
	attrs.Bold = value
	if value {
		attrs.explicitZero &^= offBold
	} else {
		attrs.explicitZero |= offBold
	}
}

func (attrs RunAttrs) BoldSpecified() bool { return attrs.Bold || attrs.explicitZero&offBold != 0 }

func (attrs *RunAttrs) SetItalic(value bool) {
	attrs.Italic = value
	if value {
		attrs.explicitZero &^= offItalic
	} else {
		attrs.explicitZero |= offItalic
	}
}

func (attrs RunAttrs) ItalicSpecified() bool {
	return attrs.Italic || attrs.explicitZero&offItalic != 0
}

func (attrs *RunAttrs) SetStrike(value bool) {
	attrs.Strike = value
	if value {
		attrs.explicitZero &^= offStrike
	} else {
		attrs.explicitZero |= offStrike
	}
}

func (attrs RunAttrs) StrikeSpecified() bool {
	return attrs.Strike || attrs.explicitZero&offStrike != 0
}

func (attrs *RunAttrs) SetSmallCaps(value bool) {
	attrs.SmallCaps = value
	if value {
		attrs.explicitZero &^= offSmallCaps
	} else {
		attrs.explicitZero |= offSmallCaps
	}
}

func (attrs RunAttrs) SmallCapsSpecified() bool {
	return attrs.SmallCaps || attrs.explicitZero&offSmallCaps != 0
}

func (attrs *RunAttrs) SetAllCaps(value bool) {
	attrs.AllCaps = value
	if value {
		attrs.explicitZero &^= offAllCaps
	} else {
		attrs.explicitZero |= offAllCaps
	}
}

func (attrs RunAttrs) AllCapsSpecified() bool {
	return attrs.AllCaps || attrs.explicitZero&offAllCaps != 0
}

type runAttrsAlias RunAttrs

func specifiedValue[T any](value T, specified bool) *T {
	if !specified {
		return nil
	}
	return &value
}

func (attrs RunAttrs) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		runAttrsAlias
		Bold      *bool           `json:"bold,omitempty"`
		Italic    *bool           `json:"italic,omitempty"`
		Strike    *bool           `json:"strike,omitempty"`
		SmallCaps *bool           `json:"small_caps,omitempty"`
		AllCaps   *bool           `json:"all_caps,omitempty"`
		Tracking  *float64        `json:"tracking,omitempty"`
		Baseline  *BaselineShift  `json:"baseline,omitempty"`
		Underline *UnderlineStyle `json:"underline,omitempty"`
		Color     *Color          `json:"color,omitempty"`
	}{runAttrsAlias: runAttrsAlias(attrs),
		Bold:      specifiedValue(attrs.Bold, attrs.BoldSpecified()),
		Italic:    specifiedValue(attrs.Italic, attrs.ItalicSpecified()),
		Strike:    specifiedValue(attrs.Strike, attrs.StrikeSpecified()),
		SmallCaps: specifiedValue(attrs.SmallCaps, attrs.SmallCapsSpecified()),
		AllCaps:   specifiedValue(attrs.AllCaps, attrs.AllCapsSpecified()),
		Tracking:  specifiedValue(attrs.Tracking, attrs.TrackingSpecified()),
		Baseline:  specifiedValue(attrs.Baseline, attrs.BaselineSpecified()),
		Underline: specifiedValue(attrs.Underline, attrs.UnderlineSpecified()),
		Color:     specifiedValue(attrs.Color, attrs.ColorSpecified()),
	})
}

func (attrs *RunAttrs) UnmarshalJSON(content []byte) error {
	decoded := runAttrsAlias(*attrs)
	input := struct {
		*runAttrsAlias
		Bold      *bool           `json:"bold"`
		Italic    *bool           `json:"italic"`
		Strike    *bool           `json:"strike"`
		SmallCaps *bool           `json:"small_caps"`
		AllCaps   *bool           `json:"all_caps"`
		Tracking  *float64        `json:"tracking"`
		Baseline  *BaselineShift  `json:"baseline"`
		Underline *UnderlineStyle `json:"underline"`
		Color     *Color          `json:"color"`
	}{runAttrsAlias: &decoded}
	if err := json.Unmarshal(content, &input); err != nil {
		return err
	}
	*attrs = RunAttrs(decoded)
	if input.Color != nil {
		attrs.SetColor(*input.Color)
	}
	if input.Bold != nil {
		attrs.SetBold(*input.Bold)
	}
	if input.Italic != nil {
		attrs.SetItalic(*input.Italic)
	}
	if input.Strike != nil {
		attrs.SetStrike(*input.Strike)
	}
	if input.SmallCaps != nil {
		attrs.SetSmallCaps(*input.SmallCaps)
	}
	if input.AllCaps != nil {
		attrs.SetAllCaps(*input.AllCaps)
	}
	if input.Underline != nil {
		attrs.SetUnderline(*input.Underline)
	}
	if input.Baseline != nil {
		attrs.SetBaseline(*input.Baseline)
	}
	if input.Tracking != nil {
		attrs.SetTracking(*input.Tracking)
	}
	return nil
}

func (base RunAttrs) Overlay(override RunAttrs) RunAttrs {
	if override.BoldSpecified() {
		base.SetBold(override.Bold)
	}
	if override.ItalicSpecified() {
		base.SetItalic(override.Italic)
	}
	if override.StrikeSpecified() {
		base.SetStrike(override.Strike)
	}
	if override.SmallCapsSpecified() {
		base.SetSmallCaps(override.SmallCaps)
	}
	if override.AllCapsSpecified() {
		base.SetAllCaps(override.AllCaps)
	}
	if override.UnderlineSpecified() {
		base.SetUnderline(override.Underline)
	}
	if override.FontSize > 0 {
		base.FontSize = override.FontSize
	}
	if override.FontName != "" {
		base.FontName = override.FontName
	}
	if override.ColorSpecified() {
		base.SetColor(override.Color)
	}
	if override.Highlight != (Color{}) {
		base.Highlight = override.Highlight
	}
	if override.ShadingPattern != "" || override.Shading != (Color{}) {
		base.Shading, base.ShadingPattern, base.ShadingForeground = override.Shading, override.ShadingPattern, override.ShadingForeground
	}
	if override.BaselineSpecified() {
		base.SetBaseline(override.Baseline)
	}
	if override.TrackingSpecified() {
		base.SetTracking(override.Tracking)
	}
	if override.Language != "" {
		base.Language = override.Language
	}
	return base
}

func (attrs *RunAttrs) SetUnderline(value UnderlineStyle) {
	attrs.Underline = value
	if value != UnderlineNone {
		attrs.explicitZero &^= offUnderline
	} else {
		attrs.explicitZero |= offUnderline
	}
}

func (attrs RunAttrs) UnderlineSpecified() bool {
	return attrs.Underline != UnderlineNone || attrs.explicitZero&offUnderline != 0
}

func (attrs *RunAttrs) SetBaseline(value BaselineShift) {
	attrs.Baseline = value
	if value != BaselineNone {
		attrs.explicitZero &^= offBaseline
	} else {
		attrs.explicitZero |= offBaseline
	}
}

func (attrs RunAttrs) BaselineSpecified() bool {
	return attrs.Baseline != BaselineNone || attrs.explicitZero&offBaseline != 0
}

func (attrs *RunAttrs) SetTracking(value float64) {
	attrs.Tracking = value
	if value != 0 {
		attrs.explicitZero &^= offTracking
	} else {
		attrs.explicitZero |= offTracking
	}
}

func (attrs RunAttrs) TrackingSpecified() bool {
	return attrs.Tracking != 0 || attrs.explicitZero&offTracking != 0
}

func (attrs *RunAttrs) InheritBold() {
	attrs.Bold = false
	attrs.explicitZero &^= offBold
}

func (attrs *RunAttrs) InheritItalic() {
	attrs.Italic = false
	attrs.explicitZero &^= offItalic
}

func (attrs *RunAttrs) InheritStrike() {
	attrs.Strike = false
	attrs.explicitZero &^= offStrike
}

func (attrs *RunAttrs) InheritSmallCaps() {
	attrs.SmallCaps = false
	attrs.explicitZero &^= offSmallCaps
}

func (attrs *RunAttrs) InheritAllCaps() {
	attrs.AllCaps = false
	attrs.explicitZero &^= offAllCaps
}

func (attrs *RunAttrs) InheritUnderline() {
	attrs.Underline = UnderlineNone
	attrs.explicitZero &^= offUnderline
}

func (attrs *RunAttrs) InheritBaseline() {
	attrs.Baseline = BaselineNone
	attrs.explicitZero &^= offBaseline
}

func (attrs *RunAttrs) InheritTracking() {
	attrs.Tracking = 0
	attrs.explicitZero &^= offTracking
}

func (attrs *RunAttrs) SetColor(value Color) {
	attrs.Color = value
	if value == (Color{}) {
		attrs.explicitZero |= offColor
	} else {
		attrs.explicitZero &^= offColor
	}
}
func (attrs RunAttrs) ColorSpecified() bool {
	return attrs.Color != (Color{}) || attrs.explicitZero&offColor != 0
}
