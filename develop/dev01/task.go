package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
	"github.com/urfave/cli/v2"
)

/*
=== Базовая задача ===

Создать программу печатающую точное время с использованием NTP библиотеки.Инициализировать как go module.
Использовать библиотеку https://github.com/beevik/ntp.
Написать программу печатающую текущее время / точное время с использованием этой библиотеки.

Программа должна быть оформлена с использованием как go module.
Программа должна корректно обрабатывать ошибки библиотеки: распечатывать их в STDERR и возвращать ненулевой код выхода в OS.
Программа должна проходить проверки go vet и golint.
*/

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "server",
				Value: "0.beevik-ntp.pool.ntp.org",
			},
		},

		Action: func(ctx *cli.Context) error {
			t, err := ntp.Time(ctx.String("server"))
			if err != nil {
				return err
			}

			fmt.Println("Current time:", t.Format(time.RFC1123))
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
		return
	}
}
