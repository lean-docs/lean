package font_test

import (
	"testing"

	// "github.com/lean-docs/lean/pkg/font"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

// --- Cluster 13: Font Management ---

// C13.1 LoadTTF loads a TrueType font file and returns a Font handle.
func TestLoadTTF(t *testing.T) {

	// f, err := font.LoadTTF("testdata/Roboto-Regular.ttf")
	// require.NoError(t, err)
	// require.NotNil(t, f)
	// assert.Equal(t, "Roboto", f.Family())
	// assert.Equal(t, font.WeightRegular, f.Weight())
	// assert.False(t, f.IsItalic())
}

// C13.2 LoadWOFF2 loads a WOFF2 web font file.
func TestLoadWOFF2(t *testing.T) {

	// f, err := font.LoadWOFF2("testdata/Roboto-Regular.woff2")
	// require.NoError(t, err)
	// require.NotNil(t, f)
	// assert.Equal(t, "Roboto", f.Family())
}

// C13.3 FontMetrics returns ascent, descent, and line gap for a font.
func TestFontMetrics(t *testing.T) {

	// f, err := font.LoadTTF("testdata/Roboto-Regular.ttf")
	// require.NoError(t, err)
	// m := f.Metrics()
	// assert.Greater(t, m.Ascent, 0.0, "ascent should be positive")
	// assert.Less(t, m.Descent, 0.0, "descent should be negative")
	// assert.GreaterOrEqual(t, m.LineGap, 0.0, "line gap should be non-negative")
	// assert.Greater(t, m.UnitsPerEm, 0, "units per em should be positive")
}

// C13.4 GlyphWidth returns the advance width for a given glyph at a specific size.
func TestGlyphWidth(t *testing.T) {

	// f, err := font.LoadTTF("testdata/Roboto-Regular.ttf")
	// require.NoError(t, err)
	// w, err := f.GlyphWidth('A', 12.0)
	// require.NoError(t, err)
	// assert.Greater(t, w, 0.0, "glyph width for 'A' should be positive")
	//
	// // Space should be narrower than 'M'
	// wSpace, _ := f.GlyphWidth(' ', 12.0)
	// wM, _ := f.GlyphWidth('M', 12.0)
	// assert.Less(t, wSpace, wM, "space should be narrower than M")
}

// C13.5 FontFallback selects a fallback font when a glyph is missing.
func TestFontFallback(t *testing.T) {

	// reg := font.NewRegistry()
	// _ = reg.LoadTTF("testdata/Roboto-Regular.ttf")
	// _ = reg.LoadTTF("testdata/NotoSansCJK-Regular.ttf")
	//
	// // Chinese character not in Roboto, should fall back to Noto
	// resolved, err := reg.Resolve("Roboto", '中')
	// require.NoError(t, err)
	// assert.Equal(t, "Noto Sans CJK", resolved.Family(),
	// 	"should fall back to CJK font for Chinese characters")
}

// C13.6 FontKerning returns kerning adjustments for character pairs.
func TestFontKerning(t *testing.T) {

	// f, err := font.LoadTTF("testdata/Roboto-Regular.ttf")
	// require.NoError(t, err)
	//
	// kern, err := f.Kern('A', 'V', 12.0)
	// require.NoError(t, err)
	// // AV is a classic kerning pair; the adjustment should be negative
	// assert.Less(t, kern, 0.0,
	// 	"AV kerning should be negative (characters drawn closer)")
	//
	// // Non-kerned pair should return zero
	// kern2, _ := f.Kern('I', 'I', 12.0)
	// assert.Equal(t, 0.0, kern2)
}

// C13.7 FontUnicode correctly handles multi-byte Unicode text measurement.
func TestFontUnicode(t *testing.T) {

	// f, err := font.LoadTTF("testdata/NotoSans-Regular.ttf")
	// require.NoError(t, err)
	//
	// // Measure an emoji (multi-byte)
	// w, err := f.StringWidth("Hello 🌍", 12.0)
	// require.NoError(t, err)
	// assert.Greater(t, w, 0.0)
	//
	// // Arabic text (RTL)
	// wArabic, err := f.StringWidth("مرحبا", 12.0)
	// require.NoError(t, err)
	// assert.Greater(t, wArabic, 0.0)
}

// C13.8 FontCache caches font data after first load to avoid repeated disk I/O.
func TestFontCache(t *testing.T) {

	// cache := font.NewCache(8) // 8 entries max
	//
	// f1, err := cache.Load("testdata/Roboto-Regular.ttf")
	// require.NoError(t, err)
	//
	// f2, err := cache.Load("testdata/Roboto-Regular.ttf")
	// require.NoError(t, err)
	//
	// // Same pointer means cache hit
	// assert.Same(t, f1, f2, "second load should return cached font")
	// assert.Equal(t, 1, cache.Len(), "cache should contain exactly one entry")
}
