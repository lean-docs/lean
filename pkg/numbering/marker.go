// Package numbering provides OOXML list numbering state and marker helpers.
package numbering

// Format is the marker style for a list level.
type Format int

const (
	Bullet Format = iota
	Decimal
	LowerLetter
	UpperLetter
	LowerRoman
	UpperRoman
)

const (
	DefaultHangingPx = 18
	MarkerGapPx      = 8
	DefaultBullet    = "•"
)
