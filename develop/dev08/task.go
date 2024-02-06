package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

/*
=== Взаимодействие с ОС ===

Необходимо реализовать собственный шелл

встроенные команды: cd/pwd/echo/kill/ps
поддержать fork/exec команды
конвеер на пайпах

Реализовать утилиту netcat (nc) клиент
принимать данные из stdin и отправлять в соединение (tcp/udp)
Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

// Command - функция для встроенной команды
type Command func(args []string, stdin io.Reader, stdout io.WriteCloser) error

type Shell struct {
	// Список встроенных команд
	commands map[string]Command

	// Ввод/вывод для шелла
	out *os.File
	in  *os.File
}

// Простейшая функция для вывода текста
func (s *Shell) printf(format string, args ...any) {
	_, _ = fmt.Fprintf(s.out, format, args...)
}

// Run запускает шелл
func (s *Shell) Run() {
	scanner := bufio.NewScanner(s.in)
	s.printf("> ")

	// Обрабатываем отдельно каждую строку как цепочку команд
	for scanner.Scan() {
		line := scanner.Text()

		// Разделяем цепочку команд на отдельные команды
		rawCommands := strings.Split(line, "|")

		// Обрабатываем цепочку команд
		err := s.processChain(rawCommands)
		if err != nil {
			s.printf("error: %s\n", err)
			continue
		}

		s.printf("> ")
	}
}

// Общий тип для команд встроенных и внешних
type executable interface {
	// setStdin устанавливает стандартный ввод для команды
	setStdin(stdin io.Reader)

	// setStdout устанавливает стандартный вывод для команды
	setStdout(stdout *os.File)

	// getStdoutPipe создаёт пайп для стандартного вывода
	getStdoutPipe() (io.ReadCloser, error)

	// exec выполняет данную команду
	exec() error
}

// ExternalCommand - внешняя команда, которая не определена в шелле, обёртка над exec.Cmd.
type ExternalCommand struct {
	cmd *exec.Cmd
}

func NewExternalCommand(name string, args ...string) *ExternalCommand {
	return &ExternalCommand{
		cmd: exec.Command(name, args...),
	}
}

func (e *ExternalCommand) setStdin(stdin io.Reader) {
	e.cmd.Stdin = stdin
}

func (e *ExternalCommand) setStdout(stdout *os.File) {
	e.cmd.Stdout = stdout
}

func (e *ExternalCommand) getStdoutPipe() (io.ReadCloser, error) {
	// Используем метод StdoutPipe, чтобы получить ридер из стандартного вывода команды
	pipe, err := e.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	return pipe, nil
}

func (e *ExternalCommand) exec() error {
	return e.cmd.Run()
}

// InternalCommand - внутренняя команда шелла, обрабатывается самим шеллом без внешних вызовов.
type InternalCommand struct {
	// Стандартный ввод/вывод для команды
	stdin  io.Reader
	stdout *os.File

	// Функция, выполняющая непосредственно логику команды
	command Command

	// Аргументы, которые будут переданы в команду
	args []string

	// Флаг, определяющий, нужно ли после выполнения команды закрыть стандартный вывод или нет.
	// Стандартный вывод ОС закрывать не стоит, а вот пайпы между командами стоит закрыть.
	shouldCloseOut bool
}

func NewInternalCommand(command Command, args ...string) *InternalCommand {
	return &InternalCommand{
		command: command,
		args:    args,
	}
}

func (i *InternalCommand) setStdin(stdin io.Reader) {
	i.stdin = stdin
}

func (i *InternalCommand) setStdout(stdout *os.File) {
	i.stdout = stdout
}

func (i *InternalCommand) getStdoutPipe() (io.ReadCloser, error) {
	// Создаём новый пайп
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, err
	}

	// Указываем, что стандартным выводом команды будет созданный выше пайп.
	// Таким образом выводимая информация будет перенаправляться в пайп, а оттуда в стандартный ввод другой команды.
	i.stdout = writer
	i.shouldCloseOut = true

	return reader, nil
}

func (i *InternalCommand) exec() error {
	// Передаём управление в функцию, отвечающую за логику данной команды
	err := i.command(i.args, i.stdin, i.stdout)
	if err != nil {
		return err
	}

	// Если нужно, закрываем стандартный вывод
	if i.shouldCloseOut {
		return i.stdout.Close()
	}

	return nil
}

// parseCommand преобразует массив параметров в команду - внутреннюю или внешнюю
func (s *Shell) parseCommand(args []string) (executable, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("no command")
	}

	// Если есть внутренняя команда с нужным нам названием, то возвращаем соответствующую структуру
	if command, ok := s.commands[args[0]]; ok {
		return NewInternalCommand(command, args[1:]...), nil
	}

	// В иных случаях считаем, что команда - внешняя
	return NewExternalCommand(args[0], args[1:]...), nil
}

// processChain обрабатывает цепочку команд
func (s *Shell) processChain(rawCommands []string) error {
	wg := &sync.WaitGroup{}
	commands := make([]executable, len(rawCommands))

	for i, rawCommand := range rawCommands {
		// Убираем лишние пробелы по краям и разделяем по оставшимся пробелам.
		// Получаем массив аргументов команды.
		args := strings.Split(strings.TrimSpace(rawCommand), " ")

		// Преобразуем аргументы в саму команду
		command, err := s.parseCommand(args)
		if err != nil {
			return err
		}

		// Для первой команды явно задаём стандартный ввод. Им будет являться стандартный ввод шелла.
		if i == 0 {
			command.setStdin(s.in)
		}

		// Для последней команды явно задаём стандартный вывод. Им будет являться стандартный вывод шелла.
		if i == len(rawCommands)-1 {
			command.setStdout(s.out)
		}

		commands[i] = command
	}

	// Проходим ещё раз по списку команд и настраиваем пайпы между ними
	for i, command := range commands[1:] {
		previousCommand := commands[i]

		// Получаем ридер из стандартного вывода предыдущей команды
		pipe, err := previousCommand.getStdoutPipe()
		if err != nil {
			return err
		}

		// Полученный ридер будет являться стандартным вводом текущей команды
		command.setStdin(pipe)

		// Стандартный ввод для первой команды и стандартный вывод для последней команды заданы при первом обходе
		// списка команд в цепочке.
	}

	// Запускаем выполнение всех команд, предварительно прибавив их количество к счётчику WaitGroup
	wg.Add(len(commands))
	for _, command := range commands {
		// Запускаем команду в отдельной горутине
		go func(cmd executable, wg *sync.WaitGroup) {
			// При завершении работы команды сообщаем об этом WaitGroup
			defer wg.Done()

			err := cmd.exec()
			if err != nil {
				fmt.Println(err)
			}
		}(command, wg)
	}

	// Ожидаем завершение выполнения всех команд
	wg.Wait()
	return nil
}

func main() {
	// Создаём шелл и определяем встроенные команды
	s := &Shell{
		commands: map[string]Command{
			"cd": func(args []string, stdin io.Reader, stdout io.WriteCloser) error {
				if len(args) < 1 {
					return nil
				}

				// Сообщаем операционной системе, что меняем рабочую директорию
				return os.Chdir(strings.Join(args, " "))
			},
			"pwd": func(args []string, stdin io.Reader, stdout io.WriteCloser) error {
				// Получаем текущую рабочую директорию
				dir, err := os.Getwd()
				if err != nil {
					return err
				}

				// Выводим полученную директорию в стандартный вывод команды
				_, err = fmt.Fprintln(stdout, dir)
				return err
			},
			"echo": func(args []string, stdin io.Reader, stdout io.WriteCloser) error {
				// Выводим в стандартный вывод команды все её аргументы через пробел
				_, err := fmt.Fprintln(stdout, strings.Join(args, " "))
				return err
			},
			"kill": func(args []string, stdin io.Reader, stdout io.WriteCloser) error {
				if len(args) < 1 {
					return fmt.Errorf("not enough arguments")
				}

				for _, strPid := range args {
					// Преобразуем i-ый аргумент команды в число
					pid, err := strconv.Atoi(strPid)
					if err != nil {
						return fmt.Errorf("illegal pid: %s", args[0])
					}

					// Просим ОС убить процесс
					err = syscall.Kill(pid, syscall.SIGKILL)
					if err != nil {
						_, _ = fmt.Fprintf(stdout, "kill %d: %s\n", pid, err)
					}
				}

				return nil
			},
		},

		// В качестве стандартного ввода/вывода шелла используем системные ввод и вывод
		out: os.Stdin,
		in:  os.Stdout,
	}

	// Запускаем шелл
	s.Run()
}
