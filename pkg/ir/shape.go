package ir

import "encoding/json"

type Shape struct {
	ID                string         `json:"id"`
	Geometry          string         `json:"geometry"`
	Width             float64        `json:"width"`
	Height            float64        `json:"height"`
	Anchor            *DrawingAnchor `json:"anchor,omitempty"`
	Fill              *Color         `json:"fill,omitempty"`
	Stroke            *Border        `json:"stroke,omitempty"`
	TextBox           bool           `json:"text_box,omitempty"`
	Padding           Padding        `json:"padding"`
	AutoFit           bool           `json:"auto_fit,omitempty"`
	VerticalAlignment string         `json:"vertical_alignment,omitempty"`
	Blocks            []Block        `json:"blocks"`
}

func (shape *Shape) BlockType() BlockType { return BlockShape }
func (shape *Shape) BlockID() string      { return shape.ID }
func (shape Shape) MarshalJSON() ([]byte, error) {
	type alias Shape
	blocks, err := MarshalBlockSlice(shape.Blocks)
	if err != nil {
		return nil, err
	}
	return json.Marshal(struct {
		alias
		Blocks json.RawMessage `json:"blocks"`
	}{alias: alias(shape), Blocks: blocks})
}
func (shape *Shape) UnmarshalJSON(data []byte) error {
	type alias Shape
	value := struct {
		*alias
		Blocks json.RawMessage `json:"blocks"`
	}{alias: (*alias)(shape)}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	blocks, err := UnmarshalBlockSlice(value.Blocks)
	if err != nil {
		return err
	}
	shape.Blocks = blocks
	return nil
}
