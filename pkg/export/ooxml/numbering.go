package ooxml

import (
	"strconv"

	"github.com/lean-docs/lean/pkg/ir"
)

func (state *packageState) prepareNumbering() {
	state.numberingIDs = make(map[*ir.Paragraph]string)
	used := map[string]bool{}
	for _, definition := range state.doc.Styles.Numbering {
		if value, err := strconv.Atoi(definition.ID); err == nil && value > 0 {
			used[definition.ID] = true
		}
	}
	next := 1
	allocate := func() string {
		for used[strconv.Itoa(next)] {
			next++
		}
		id := strconv.Itoa(next)
		used[id] = true
		next++
		return id
	}
	mapped := map[string]string{}
	definitions := map[string]ir.NumberingDef{}
	for _, definition := range state.doc.Styles.Numbering {
		copy := definition
		if value, err := strconv.Atoi(copy.ID); err != nil || value <= 0 {
			copy.ID = allocate()
		}
		mapped[definition.ID] = copy.ID
		definitions[definition.ID] = copy
		state.numbering = append(state.numbering, copy)
	}
	active := map[string]string{}
	visit := func(block ir.Block) {
		paragraph, ok := block.(*ir.Paragraph)
		if !ok || paragraph.Numbering == nil {
			return
		}
		reference := paragraph.Numbering
		id := active[reference.ID]
		if id == "" {
			id = mapped[reference.ID]
		}
		if reference.Restart != nil {
			definition := definitions[reference.ID]
			definition.ID = allocate()
			definition.Levels = append([]ir.NumberingLevel(nil), definition.Levels...)
			for index := range definition.Levels {
				if definition.Levels[index].Level == reference.Level {
					definition.Levels[index].Start = *reference.Restart
				}
			}
			state.numbering = append(state.numbering, definition)
			id = definition.ID
			active[reference.ID] = id
		}
		state.numberingIDs[paragraph] = id
	}
	for _, section := range state.doc.Sections {
		walkBlocks(section.Blocks, visit)
	}
	for _, part := range state.auxiliary {
		walkBlocks(part.value.Blocks, visit)
	}
	for _, note := range state.notes {
		walkBlocks(note.Blocks, visit)
	}
}
