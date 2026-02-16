package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	cursorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
)

type model struct {
	newFileInput           textinput.Model
	createFileInputVisible bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit
		case "ctrl+n":
			m.createFileInputVisible = true
			return m, nil
		}
	}

	if m.createFileInputVisible {
		m.newFileInput, cmd = m.newFileInput.Update(msg)
	}

	return m, cmd
}

func (m model) View() string {

	var style = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("16")).
		Background(lipgloss.Color("205")).
		PaddingLeft(2).
		PaddingRight(2)

	welcome := style.Render("welcome to tuigote :x")
	help := "Ctrl+N:new file | Ctrl+S:save | Ctrl+L:list | Esc:back/save | Ctrl+Q:quit"
	view := " "
	if m.createFileInputVisible {
		view = m.newFileInput.View()
	}

	return fmt.Sprintf("\n%s\n\n%s\n\n%s", welcome, view, help)
}

func initializeMode() model {

	//initialize new file input
	ti := textinput.New()
	ti.Placeholder = "Name your file?"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20
	ti.Cursor.Style = cursorStyle
	ti.PromptStyle = cursorStyle

	return model{
		newFileInput:           ti,
		createFileInputVisible: false,
	}
}

func main() {
	p := tea.NewProgram(initializeMode())
	if _, err := p.Run(); err != nil {
		fmt.Printf("there's been an error goon: %v", err)
		os.Exit(1)
	}
}
