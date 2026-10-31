package lean

import (
	"crypto/rand"
	"errors"
	"slices"

	"github.com/lean-docs/lean/pkg/ir"
)

type DrawingLayout struct {
	Width  *float64          `json:"width,omitempty"`
	Height *float64          `json:"height,omitempty"`
	Anchor *ir.DrawingAnchor `json:"anchor,omitempty"`
}

func applyDrawingCommand(blocks *[]ir.Block, command DocumentCommand) error {
	if command.ID == "image.insert" || command.ID == "shape.insert" {
		if command.At < 0 || command.At > len(*blocks) {
			return errors.New("invalid insertion position")
		}
		var drawing ir.Block
		if command.ID == "image.insert" {
			if command.Image == nil {
				return errors.New("an image is required")
			}
			if command.Image.ID == "" {
				command.Image.ID = rand.Text()
			}
			drawing = command.Image
		} else {
			if command.Shape == nil {
				return errors.New("a shape is required")
			}
			if command.Shape.ID == "" {
				command.Shape.ID = rand.Text()
			}
			drawing = command.Shape
		}
		*blocks = slices.Insert(*blocks, command.At, drawing)
		return nil
	}
	parent, index, err := findCommandBlock(blocks, command.Block)
	if err != nil {
		return err
	}
	var width, height *float64
	var anchor **ir.DrawingAnchor
	switch drawing := (*parent)[index].(type) {
	case *ir.Image:
		width, height, anchor = &drawing.Width, &drawing.Height, &drawing.Anchor
	case *ir.Shape:
		width, height, anchor = &drawing.Width, &drawing.Height, &drawing.Anchor
	default:
		return errors.New("an image or shape is required")
	}
	switch command.ID {
	case "drawing.delete":
		*parent = slices.Delete(*parent, index, index+1)
	case "drawing.layout":
		if command.Drawing == nil {
			return errors.New("a drawing layout is required")
		}
		if command.Drawing.Width != nil {
			*width = *command.Drawing.Width
		}
		if command.Drawing.Height != nil {
			*height = *command.Drawing.Height
		}
		if command.Drawing.Anchor != nil {
			*anchor = command.Drawing.Anchor
		}
	default:
		return errors.New("unknown drawing command")
	}
	return nil
}
