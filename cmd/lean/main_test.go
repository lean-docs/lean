package main

import (
	"encoding/json"
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestCLICommandsApplyAndInspectCanonicalDOCX(t *testing.T) {
	directory := t.TempDir()
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{ID: "intro", Runs: []ir.Run{{Text: "Original"}}}}
	content, _, err := lean.SaveDOCX(document)
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(directory, "input.docx")
	if err := os.WriteFile(input, content, 0o600); err != nil {
		t.Fatal(err)
	}
	reopened, _, err := lean.OpenDOCX(content)
	if err != nil {
		t.Fatal(err)
	}
	commands := []lean.DocumentCommand{{ID: "text.replace", Block: reopened.Sections[0].Blocks[0].BlockID(), Start: 0, End: 8, Text: "Updated"}, {ID: "paragraph.style", Block: reopened.Sections[0].Blocks[0].BlockID(), Style: "Heading1"}}
	encoded, err := json.Marshal(commands)
	if err != nil {
		t.Fatal(err)
	}
	instructions := filepath.Join(directory, "commands.json")
	if err := os.WriteFile(instructions, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(directory, "output.docx")
	command := exec.Command("go", "run", ".", "docx", "apply", input, instructions, output)
	if result, err := command.CombinedOutput(); err != nil {
		t.Fatalf("apply: %v %s", err, result)
	}
	command = exec.Command("go", "run", ".", "docx", "inspect", output)
	result, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("inspect: %v %s", err, result)
	}
	var inspected struct {
		Document ir.Document
		Fidelity lean.FidelityReport
	}
	if err := json.Unmarshal(result, &inspected); err != nil {
		t.Fatal(err)
	}
	paragraph := inspected.Document.Sections[0].Blocks[0].(*ir.Paragraph)
	if paragraph.Style != "Heading1" || paragraph.Runs[0].Text != "Updated" || !inspected.Fidelity.Editable {
		t.Fatal("CLI output did not retain the requested edit")
	}
	if result, err := exec.Command("go", "run", ".", "docx", "apply", input, instructions, output).CombinedOutput(); err == nil {
		t.Fatalf("CLI overwrote an existing output: %s", result)
	}
	original, err := os.ReadFile(input)
	if err != nil || string(original) != string(content) {
		t.Fatal("CLI changed its input document")
	}
}
