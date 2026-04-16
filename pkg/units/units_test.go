package units

import (
	"math"
	"testing"
)

func TestPixelsToTwips(t *testing.T) {
	cases := []struct {
		in   float64
		want int
	}{
		{96, 1440},
		{24, 360},
		{0, 0},
	}
	for _, c := range cases {
		if got := PixelsToTwips(c.in); got != c.want {
			t.Errorf("PixelsToTwips(%v) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestPixelsToTwipsInvalid(t *testing.T) {
	if got := PixelsToTwips(math.NaN()); got != 0 {
		t.Errorf("NaN → %d, want 0", got)
	}
	if got := PixelsToTwips(math.Inf(1)); got != 0 {
		t.Errorf("Inf → %d, want 0", got)
	}
}

func TestTwipsRoundTripPoints(t *testing.T) {
	for _, pt := range []float64{12, 14.5, 72, 0.1} {
		back := TwipsToPoints(PointsToTwips(pt))
		if math.Abs(back-pt) > 1e-9 {
			t.Errorf("roundtrip %v → %v", pt, back)
		}
	}
}

func TestHalfPoints(t *testing.T) {
	if HalfPointsToPoints(24) != 12 {
		t.Fatal("24 half-points should be 12 points")
	}
	if PointsToHalfPoints(12) != 24 {
		t.Fatal("12 points should be 24 half-points")
	}
	if HalfPointsToPoints(5) != 2.5 {
		t.Fatal("5 half-points should be 2.5 points")
	}
}

func TestEMU(t *testing.T) {
	if PointsToEMU(72) != EMUPerInch {
		t.Errorf("72pt = 1in = %d EMU, got %d", EMUPerInch, PointsToEMU(72))
	}
	if EMUToPoints(EMUPerInch) != 72 {
		t.Error("1in EMU should be 72pt")
	}
	if got := EMUToPixels(EMUPerInch); got != 96 {
		t.Errorf("1in EMU in pixels = %v, want 96", got)
	}
}

func TestInvalidInputsReturnZero(t *testing.T) {
	bad := math.NaN()
	if PointsToTwips(bad) != 0 || TwipsToPoints(bad) != 0 ||
		TwipsToPixels(bad) != 0 || HalfPointsToPoints(bad) != 0 ||
		PointsToHalfPoints(bad) != 0 || PointsToEMU(bad) != 0 {
		t.Fatal("NaN should produce 0 across all converters")
	}
}
