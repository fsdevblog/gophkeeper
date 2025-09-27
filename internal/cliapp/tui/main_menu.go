package tui

import (
	"fmt"
	"strings"
)

type MainMenuChoice int

const (
	MMChoiceLogin MainMenuChoice = iota
	MMChoiceRegister
)

type MainMenu struct {
	choices  []string
	selected MainMenuChoice
}

func NewMainMenu() *MainMenu {
	return &MainMenu{
		choices:  []string{"Login", "Register"},
		selected: MMChoiceLogin,
	}
}

func (m *MainMenu) Next() {
	selected := (int(m.selected) + 1) % len(m.choices)
	m.selected = MainMenuChoice(selected)
}

func (m *MainMenu) Prev() {
	selected := (int(m.selected) - 1 + len(m.choices)) % len(m.choices)
	m.selected = MainMenuChoice(selected)
}

func (m *MainMenu) GetSelected() MainMenuChoice {
	return m.selected
}

func (m *MainMenu) Render() string {
	var b strings.Builder
	b.WriteString("Select action: \n")

	for i, choice := range m.choices {
		cursor := " "
		if i == int(m.selected) {
			cursor = ">"
		}
		b.WriteString(fmt.Sprintf("%s %s\n", cursor, choice))
	}
	b.WriteString("\nuse ↑/↓ for navigation, enter to select")
	return b.String()
}
