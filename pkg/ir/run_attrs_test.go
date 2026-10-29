package ir_test

import (
	"encoding/json"
	"github.com/lean-docs/lean/pkg/ir"
	"testing"
)

func TestRunFormattingDistinguishesInheritedOnAndOff(t *testing.T) {
	for _, test := range []struct {
		name      string
		set       func(*ir.RunAttrs, bool)
		specified func(ir.RunAttrs) bool
		inherit   func(*ir.RunAttrs)
	}{
		{"bold", (*ir.RunAttrs).SetBold, ir.RunAttrs.BoldSpecified, (*ir.RunAttrs).InheritBold},
		{"italic", (*ir.RunAttrs).SetItalic, ir.RunAttrs.ItalicSpecified, (*ir.RunAttrs).InheritItalic},
		{"strike", (*ir.RunAttrs).SetStrike, ir.RunAttrs.StrikeSpecified, (*ir.RunAttrs).InheritStrike},
		{"small_caps", (*ir.RunAttrs).SetSmallCaps, ir.RunAttrs.SmallCapsSpecified, (*ir.RunAttrs).InheritSmallCaps},
		{"all_caps", (*ir.RunAttrs).SetAllCaps, ir.RunAttrs.AllCapsSpecified, (*ir.RunAttrs).InheritAllCaps},
	} {
		t.Run(test.name, func(t *testing.T) {
			var attrs ir.RunAttrs
			if test.specified(attrs) {
				t.Fatal("zero value must inherit")
			}
			for _, enabled := range []bool{false, true, false} {
				test.set(&attrs, enabled)
				content, err := json.Marshal(attrs)
				if err != nil {
					t.Fatal(err)
				}
				var encoded map[string]any
				if err := json.Unmarshal(content, &encoded); err != nil {
					t.Fatal(err)
				}
				if value, present := encoded[test.name]; !present || value != enabled {
					t.Fatalf("lost explicit value: %s", content)
				}
				var restored ir.RunAttrs
				if err := json.Unmarshal(content, &restored); err != nil {
					t.Fatal(err)
				}
				if !test.specified(restored) || restored != attrs {
					t.Fatalf("lost override: %s", content)
				}
			}
			attrs.FontName = "Arial"
			test.inherit(&attrs)
			if test.specified(attrs) || attrs.FontName != "Arial" {
				t.Fatal("inheritance reset changed unrelated formatting")
			}
			encoded, err := json.Marshal(attrs)
			if err != nil {
				t.Fatal(err)
			}
			var properties map[string]any
			if err := json.Unmarshal(encoded, &properties); err != nil {
				t.Fatal(err)
			}
			if _, present := properties[test.name]; present {
				t.Fatalf("inherited property remains explicit: %s", encoded)
			}
		})
	}
}

func TestParagraphStyleRunInheritanceAndDirectOverrides(t *testing.T) {
	off := ir.RunAttrs{Strike: true, Tracking: 1.5, Language: "fr-FR"}
	off.SetBold(false)
	sheet := ir.StyleSheet{
		Defaults: ir.RunAttrs{FontName: "Arial", Italic: true},
		Named: map[string]ir.Style{
			"Parent": {ID: "Parent", RunAttrs: ir.RunAttrs{Bold: true, FontSize: 20, Underline: ir.UnderlineDouble}},
			"Child":  {ID: "Child", BasedOn: "Parent", RunAttrs: off},
		},
	}
	style, ok := sheet.ResolveStyle("Child")
	if !ok || style.RunAttrs.Bold || !style.RunAttrs.BoldSpecified() || !style.RunAttrs.Italic || style.RunAttrs.FontName != "Arial" || style.RunAttrs.FontSize != 20 || style.RunAttrs.Underline != ir.UnderlineDouble || !style.RunAttrs.Strike || style.RunAttrs.Tracking != 1.5 || style.RunAttrs.Language != "fr-FR" {
		t.Fatalf("incorrect inheritance: %+v", style)
	}
	var direct ir.RunAttrs
	direct.SetItalic(false)
	result := style.RunAttrs.Overlay(direct)
	if result.Italic || !result.ItalicSpecified() || result.FontSize != 20 {
		t.Fatalf("incorrect direct override: %+v", result)
	}
}

func TestStyleCyclesFailWithoutRecursing(t *testing.T) {
	sheet := ir.StyleSheet{Named: map[string]ir.Style{
		"First":  {ID: "First", BasedOn: "Second"},
		"Second": {ID: "Second", BasedOn: "First"},
	}}
	if _, ok := sheet.ResolveStyle("First"); ok {
		t.Fatal("cyclic styles must not resolve")
	}
}

func TestRunValueInheritanceResets(t *testing.T) {
	var attrs ir.RunAttrs
	attrs.SetUnderline(ir.UnderlineNone)
	attrs.SetBaseline(ir.BaselineNone)
	attrs.SetTracking(0)
	attrs.SetBold(false)
	attrs.InheritUnderline()
	attrs.InheritBaseline()
	attrs.InheritTracking()
	if attrs.UnderlineSpecified() || attrs.BaselineSpecified() || attrs.TrackingSpecified() || !attrs.BoldSpecified() {
		t.Fatal("inheritance reset changed unrelated explicit formatting")
	}
	resolved := (ir.RunAttrs{Underline: ir.UnderlineDouble, Baseline: ir.BaselineSuperscript, Tracking: 2}).Overlay(attrs)
	if resolved.Underline != ir.UnderlineDouble || resolved.Baseline != ir.BaselineSuperscript || resolved.Tracking != 2 {
		t.Fatalf("parent formatting did not resume: %+v", resolved)
	}
}
