package views

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsdevblog/gophkeeper/internal/cliapp/tui"
	"strings"
)

var _ tea.Model = &PublicMenuModel{}

type PublicMenuChoice string

const (
	PublicMenuLogin    PublicMenuChoice = "login"
	PublicMenuRegister PublicMenuChoice = "register"
)

type PublicMenuModel struct {
	choices []PublicMenuChoice
	choice  PublicMenuChoice
	cursor  int
}

func NewPublicMenu() *PublicMenuModel {

	return &PublicMenuModel{
		choices: []PublicMenuChoice{PublicMenuLogin, PublicMenuRegister},
		choice:  PublicMenuLogin,
		cursor:  0,
	}
}

func (p PublicMenuModel) Init() tea.Cmd {
	return nil
}

func (p PublicMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typedMsg := msg.(type) {
	case tea.KeyMsg:
		switch typedMsg.Type {
		case tea.KeyCtrlQ, tea.KeyCtrlC:
			return p, tea.Quit
		case tea.KeyEnter:
			p.choice = p.choices[p.cursor]
			switch p.choice {
			case PublicMenuLogin:
				return p, tui.ChangeStateCmd(tui.StateLogin)
			}
		case tea.KeyDown, tea.KeyShiftTab:
			p.cursor++
			if p.cursor >= len(p.choices) {
				p.cursor = 0
			}
		case tea.KeyUp, tea.KeyTab:
			p.cursor--
			if p.cursor < 0 {
				p.cursor = len(p.choices) - 1
			}
		}
	}
	return p, nil
}

func (p PublicMenuModel) View() string {
	b := new(strings.Builder)
	b.WriteString("You need to login or register to continue.\n")
	for i, choice := range p.choices {
		if p.cursor == i {
			b.WriteString("(•) ")
		} else {
			b.WriteString("( ) ")
		}
		b.WriteString(string(choice))
		b.WriteRune('\n')
	}
	b.WriteString("\npress ctrl+q to quit\n")
	return b.String()
}
