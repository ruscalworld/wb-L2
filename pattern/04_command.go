package pattern

import "errors"

/*
	Реализовать паттерн «комманда».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Command_pattern

Паттерн подразумевает выделение отдельных структур под отдельные действия.

Преимущества - возможность инкапсулировать данные, необходимые для отдельного вызова, в некоторой структуре, которая
будет отвечать за выполнение данного вызова; возможность вызывать определённые действия, зная лишь
некоторый их идентификатор, не задумываясь о конкретной реализации.

Пример - удалённое выполнение какой-либо процедуры на сервере: клиенту достаточно знать лишь название процедуры и, в
некоторых случаях, информацию о принимаемых ей параметрах. Тогда клиент может передать серверу название процедуры, не
зная, как именно она реализована на сервере.
*/

// Command - общий тип для команд.
type Command interface {
	Execute() error
}

// CommandTarget содержит некоторые методы, которые будут вызываться командами.
type CommandTarget struct{}

func (*CommandTarget) SomeAction1() error {
	return nil
}

func (*CommandTarget) SomeAction2() error {
	return nil
}

// Command1 - команда, выполняющая SomeAction1.
type Command1 struct {
	target *CommandTarget
}

// NewCommand1 - конструктор для команды.
func NewCommand1(target *CommandTarget) *Command1 {
	return &Command1{target: target}
}

func (c Command1) Execute() error {
	return c.target.SomeAction1()
}

// Command2 - команда, выполняющая SomeAction1.
type Command2 struct {
	target *CommandTarget
}

// NewCommand2 - конструктор для команды.
func NewCommand2(target *CommandTarget) *Command2 {
	return &Command2{target: target}
}

func (c Command2) Execute() error {
	return c.target.SomeAction1()
}

// CommandDispatcher хранит в себе информацию о командах и позволяет их вызывать, используя их номер.
type CommandDispatcher struct {
	commands map[int]Command
}

func NewCommandDispatcher(target *CommandTarget) *CommandDispatcher {
	return &CommandDispatcher{
		commands: map[int]Command{
			1: NewCommand1(target),
			2: NewCommand2(target),
		},
	}
}

func (d *CommandDispatcher) Execute(n int) error {
	cmd, ok := d.commands[n]
	if !ok {
		return errors.New("unknown command")
	}

	return cmd.Execute()
}
