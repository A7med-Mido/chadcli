package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/A7med-Mido/chadcli/internal/ui"
)

func main() {
	// Start from the current working directory, or from the argument if provided
	startDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error getting working directory: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) > 1 {
		startDir = os.Args[1]
	}

	model, err := ui.NewModel(startDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing explorer: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error running program: %v\n", err)
		os.Exit(1)
	}
}