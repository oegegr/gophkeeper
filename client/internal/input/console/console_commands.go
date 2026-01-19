package console

import (
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/oegegr/gophkeeper/client/internal/domain"
)

func registerCommands(app *cli.App, handler *ConsoleHandler) {
	app.Commands = []*cli.Command{
		// Команда входа
		{
			Name:  login,
			Usage: "Войти в систему",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"u"},
					Usage:    "Имя пользователя",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "password",
					Aliases:  []string{"p"},
					Usage:    "Пароль",
					Required: true,
				},
			},
			Action: func(ctx *cli.Context) error {
				return handler.HandleLogin(ctx)
			},
		},

		// Команда регистрации
		{
			Name:  register,
			Usage: "Зарегистрировать нового пользователя",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "username",
					Aliases:  []string{"u"},
					Usage:    "Имя пользователя",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "password",
					Aliases:  []string{"p"},
					Usage:    "Пароль",
					Required: true,
				},
			},
			Action: func(ctx *cli.Context) error {
				return handler.HandleRegister(ctx)
			},
		},

		// Команды для работы с секретами
		{
			Name:  secret,
			Usage: "Управление секретами",
			Subcommands: []*cli.Command{
				// Список секретов
				{
					Name:  secret_list,
					Usage: "Показать все секреты",
					Action: func(ctx *cli.Context) error {
						return handler.HandleListSecrets(ctx)
					},
				},

				// Получить секрет
				{
					Name:  secret_get,
					Usage: "Получить секрет по ID",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "id",
							Aliases:  []string{"i"},
							Usage:    "ID секрета",
							Required: true,
						},
					},
					Action: func(ctx *cli.Context) error {
						return handler.HandleGetSecret(ctx)
					},
				},

				// Добавить секрет
				{
					Name:  secret_add,
					Usage: "Добавить новый секрет",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "name",
							Aliases:  []string{"n"},
							Usage:    "Название секрета",
							Required: true,
						},
						&cli.StringFlag{
							Name:     "data",
							Aliases:  []string{"d"},
							Usage:    "Данные секрета",
							Required: true,
						},
						&cli.StringFlag{
							Name:    "type",
							Aliases: []string{"t"},
							Usage:   "Тип секрета",
							Value:   "password",
						},
					},
					Action: func(ctx *cli.Context) error {
						return handler.HandleAddSecret(ctx)
					},
				},

				// Удалить секрет
				{
					Name:  secret_delete,
					Usage: "Удалить секрет",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "id",
							Aliases:  []string{"i"},
							Usage:    "ID секрета",
							Required: true,
						},
					},
					Action: func(ctx *cli.Context) error {
						// Подтверждение
						confirm := false
						if ctx.Bool("force") {
							confirm = true
						} else {
							fmt.Printf("Удалить секрет %s? (y/N): ", ctx.String("id"))
							var answer string
							fmt.Scanln(&answer)
							confirm = (answer == "y" || answer == "Y")
						}

						if !confirm {
							fmt.Println("Отменено")
							return nil
						}

						return handler.HandleDeleteSecret(ctx)
					},
				},
			},
		},

		// Команда синхронизации
		{
			Name:  sync,
			Usage: "Синхронизировать данные с сервером",
			Action: func(ctx *cli.Context) error {
				return handler.HandleForceSync(ctx)
			},
		},

		// Команда версии
		{
			Name:  version,
			Usage: "Показать версию приложения",
			Action: func(ctx *cli.Context) error {
				fmt.Println("GophKeeper v1.0.0")
				return nil
			},
		},
	}
}

// Вспомогательные методы для вывода
func ShowSecretsList(secrets []domain.Secret) {
	if len(secrets) == 0 {
		fmt.Println("Секретов нет")
		return
	}

	fmt.Println("\nВаши секреты:")
	for i, s := range secrets {
		fmt.Printf("%d. %s [%s] - %s\n",
			i+1,
			s.ID,
			s.Type,
			s.UpdatedAt.Format("2006-01-02 15:04"),
		)
	}
}