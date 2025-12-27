package views

import (
	tea "github.com/charmbracelet/bubbletea"
)

var _ tea.Model = &PrivateMenuModel{}

type PrivateMenuModel struct {
}

func NewPrivateMenu() *PrivateMenuModel {
	return &PrivateMenuModel{}
}
func (p *PrivateMenuModel) Init() tea.Cmd {
	return nil
}

func (p *PrivateMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyCtrlQ, tea.KeyCtrlC:
			return p, tea.Quit
		}
	}
	return p, nil
}

func (p *PrivateMenuModel) View() string {
	return "private menu"
}
