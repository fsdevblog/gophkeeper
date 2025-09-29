package tui

import (
	"strings"

	"github.com/google/uuid"
)

type User struct {
	Username string
	ID       uuid.UUID
}
type AuthorizedMain struct {
	user *User
}

func NewAuthorizedMain(user *User) *AuthorizedMain {
	return &AuthorizedMain{
		user: user,
	}
}

func (a *AuthorizedMain) Render() string {
	var b strings.Builder
	b.WriteString("Welcome, ")
	b.WriteString(primaryTextStyle.Render(a.user.Username) + "\n\n")
	return b.String()
}
