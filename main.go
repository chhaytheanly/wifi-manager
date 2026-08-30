package main

import (
	"fmt"
	"os"
	"wifi-tui/app"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	initialModel := app.NewModel()

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
