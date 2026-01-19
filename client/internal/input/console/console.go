package console

import (
	"github.com/urfave/cli/v2"
)

// консольные команды
const (
	register      = "register"
	login         = "login"
	version       = "version"
	help          = "help"
	secret        = "secret"
	sync          = "sync"
	secret_add    = "add"
	secret_get    = "get"
	secret_delete = "delete"
	secret_list   = "list"
)

// New создает новый консольный интерфейс
func New(handler *ConsoleHandler) (*cli.App, error) {
	app := &cli.App{
		Name:    "gophkeeper",
		Version: "1.0.0",
		Usage:   "Secure password manager",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server",
				Value: "localhost:8080",
				Usage: "Server address",
			},
		},
	}

	registerCommands(app, handler)

	return app, nil
}
