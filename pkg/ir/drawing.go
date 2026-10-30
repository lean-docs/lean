package ir

type DrawingPosition struct {
	Relative string  `json:"relative"`
	Offset   float64 `json:"offset"`
	Align    string  `json:"align,omitempty"`
}

type DrawingPolygon struct {
	Edited bool       `json:"edited,omitempty"`
	Points [][2]int64 `json:"points"`
}

type DrawingAnchor struct {
	Polygon        *DrawingPolygon `json:"polygon,omitempty"`
	ParagraphID    string          `json:"paragraph_id,omitempty"`
	Offset         int             `json:"offset"`
	Floating       bool            `json:"floating,omitempty"`
	Horizontal     DrawingPosition `json:"horizontal"`
	Vertical       DrawingPosition `json:"vertical"`
	Wrap           string          `json:"wrap,omitempty"`
	WrapSide       string          `json:"wrap_side,omitempty"`
	Distance       Padding         `json:"distance"`
	Behind         bool            `json:"behind,omitempty"`
	AllowOverlap   bool            `json:"allow_overlap,omitempty"`
	LayoutInCell   bool            `json:"layout_in_cell,omitempty"`
	ZOrder         uint32          `json:"z_order,omitempty"`
	Rotation       float64         `json:"rotation,omitempty"`
	FlipHorizontal bool            `json:"flip_horizontal,omitempty"`
	FlipVertical   bool            `json:"flip_vertical,omitempty"`
	Crop           Padding         `json:"crop"`
}
