package main

import (
	"github.com/fsdevblog/gophkeeper/internal/cliapp"
)

func main() {
	app := cliapp.New(cliapp.InitParams{BaseURL: "http://127.0.0.1:8080"})
	if err := app.Run(); err != nil {
		panic(err)
	}
}
