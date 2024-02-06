package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

/*
=== Утилита telnet ===

Реализовать примитивный telnet клиент:
Примеры вызовов:
go-telnet --timeout=10s host port go-telnet mysite.ru 8080 go-telnet --timeout=3s 1.1.1.1 123

Программа должна подключаться к указанному хосту (ip или доменное имя) и порту по протоколу TCP.
После подключения STDIN программы должен записываться в сокет, а данные полученные и сокета должны выводиться в STDOUT
Опционально в программу можно передать таймаут на подключение к серверу (через аргумент --timeout, по умолчанию 10s).

При нажатии Ctrl+D программа должна закрывать сокет и завершаться. Если сокет закрывается со стороны сервера, программа должна также завершаться.
При подключении к несуществующему сервер, программа должна завершаться через timeout.
*/

var rawTimeout = flag.String("timeout", "10s", "timeout")

func initSession(timeout time.Duration, address string) error {
	// Подключаемся к серверу
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return err
	}

	// Сообщаем об успешном подключении и создаём контекст
	fmt.Println("Connected to", address)
	ctx, cancel := context.WithCancel(context.Background())

	// Запускаем два "пайпа" для перенаправления данных из stdin в соединение и из соединения в stdout
	go pipe(ctx, cancel, conn, os.Stdin)
	go pipe(ctx, cancel, os.Stdout, conn)

	// Ожидаем завершения работы пайпов
	<-ctx.Done()
	return nil
}

func pipe(ctx context.Context, cancel context.CancelFunc, dst io.Writer, src io.Reader) {
	// Будем использовать reader из bufio
	reader := bufio.NewReader(src)

	// Если что-то случится с пайпом, сообщаем об этом наверх и через контекст завершаем работу всей программы
	defer cancel()

	for {
		select {
		// Обрабатываем сигнал завершения работы программы
		case <-ctx.Done():
			return

		// Если сигнала о завершении работы нет, то продолжаем работу
		default:
			// Читаем строку до переноса
			line, err := reader.ReadString('\n')
			if err != nil {
				// Здесь может возникнуть io.EOF, спровоцированный нажатием Ctrl+D, он будет обработан так же, как и
				// остальные ошибки.
				cancel()
				return
			}

			// Отправляем считанную строку в writer
			_, err = dst.Write([]byte(line))
			if err != nil {
				cancel()
				return
			}
		}
	}
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		fmt.Println("not enough arguments")
		return
	}

	// Преобразуем строку, полученную через флаг --timeout в time.Duration.
	timeout, err := time.ParseDuration(*rawTimeout)
	if err != nil {
		fmt.Println("invalid timeout:", err)
		return
	}

	// Формируем адрес из аргументов команды (первые два аргумента, разделённые через ":" = адрес:порт)
	address := strings.Join(flag.Args()[:2], ":")

	// Инициализируем сессию
	err = initSession(timeout, address)
	if err != nil {
		fmt.Println("error initiating session:", err)
		return
	}
}
