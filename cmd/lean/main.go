package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lean-docs/lean"
	"io"
	"os"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) == 0 || len(args) == 1 && args[0] == "--version" {
		_, err := fmt.Fprintf(output, "lean %s\n", lean.Version())
		return err
	}
	if len(args) == 2 && args[0] == "docx" && args[1] == "commands" {
		return json.NewEncoder(output).Encode(lean.DOCXCommands())
	}
	if len(args) == 3 && args[0] == "docx" && args[1] == "inspect" {
		content, err := os.ReadFile(args[2])
		if err != nil {
			return err
		}
		document, fidelity, err := lean.OpenDOCX(content)
		if err != nil {
			return err
		}
		return json.NewEncoder(output).Encode(struct {
			Document any                 `json:"document"`
			Fidelity lean.FidelityReport `json:"fidelity"`
		}{document, fidelity})
	}
	if len(args) != 5 || args[0] != "docx" || args[1] != "apply" {
		return errors.New("usage: lean docx commands | lean docx inspect INPUT | lean docx apply INPUT COMMANDS.json OUTPUT")
	}
	content, err := os.ReadFile(args[2])
	if err != nil {
		return err
	}
	document, fidelity, err := lean.OpenDOCX(content)
	if err != nil {
		return err
	}
	if !fidelity.Editable {
		return fmt.Errorf("%w: %v", lean.ErrUnsupportedDocument, fidelity.Unsupported)
	}
	file, err := os.Open(args[3])
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(io.LimitReader(file, 8<<20))
	decoder.DisallowUnknownFields()
	var commands []lean.DocumentCommand
	if err := decoder.Decode(&commands); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("the command file must contain one JSON array")
	}
	edited, err := lean.ApplyDOCXCommands(document, commands)
	if err != nil {
		return err
	}
	binary, fidelity, err := lean.SaveDOCX(edited)
	if err != nil {
		return err
	}
	destination, err := os.OpenFile(args[4], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := destination.Write(binary); err != nil {
		destination.Close()
		return err
	}
	if err := destination.Close(); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(fidelity)
}
