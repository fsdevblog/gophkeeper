package tui

type State int

const (
	StatePublicMenu = State(iota)
	StatePrivateMenu
	StateLogin
	StateRegister
)

var stateToStrMap = map[State]string{
	StatePublicMenu:  "Menu",
	StatePrivateMenu: "Menu",
	StateLogin:       "Login",
	StateRegister:    "Register",
}

func (s State) String() string {
	return stateToStrMap[s]
}
