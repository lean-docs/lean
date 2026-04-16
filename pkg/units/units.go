// Package units converts between OOXML and layout measurement units.
//
// OOXML stores most paragraph and run measurements in twips (1/20 pt) and
// half-points, and drawing measurements in EMU (English Metric Units).
// Rendering pipelines work in points or CSS pixels. These helpers keep the
// conversions consistent and handle invalid input (NaN, Inf) uniformly.
package units

import "math"

const (
	TwipsPerInch  = 1440
	PointsPerInch = 72
	PixelsPerInch = 96
	EMUPerInch    = 914400

	TwipsPerPoint = 20
	EMUPerPoint   = EMUPerInch / PointsPerInch // 12700
	TwipsPerPixel = TwipsPerInch / PixelsPerInch // 15
)

func finite(v float64) (float64, bool) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, false
	}
	return v, true
}

// PixelsToTwips rounds to the nearest twip. Invalid input → 0.
func PixelsToTwips(px float64) int {
	v, ok := finite(px)
	if !ok {
		return 0
	}
	return int(math.Round(v * float64(TwipsPerPixel)))
}

// TwipsToPixels is the inverse of PixelsToTwips without rounding.
func TwipsToPixels(twips float64) float64 {
	v, ok := finite(twips)
	if !ok {
		return 0
	}
	return v / float64(TwipsPerPixel)
}

func PointsToTwips(pt float64) float64 {
	v, ok := finite(pt)
	if !ok {
		return 0
	}
	return v * float64(TwipsPerPoint)
}

func TwipsToPoints(twips float64) float64 {
	v, ok := finite(twips)
	if !ok {
		return 0
	}
	return v / float64(TwipsPerPoint)
}

func HalfPointsToPoints(hp float64) float64 {
	v, ok := finite(hp)
	if !ok {
		return 0
	}
	return v / 2
}

func PointsToHalfPoints(pt float64) float64 {
	v, ok := finite(pt)
	if !ok {
		return 0
	}
	return v * 2
}

// EMU conversions — OOXML DrawingML stores dimensions in EMU.
func EMUToPoints(emu int64) float64 { return float64(emu) / float64(EMUPerPoint) }
func PointsToEMU(pt float64) int64 {
	v, ok := finite(pt)
	if !ok {
		return 0
	}
	return int64(math.Round(v * float64(EMUPerPoint)))
}
func EMUToPixels(emu int64) float64 { return float64(emu) / float64(EMUPerInch) * float64(PixelsPerInch) }
