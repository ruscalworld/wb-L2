package pattern

/*
	Реализовать паттерн «фасад».
Объяснить применимость паттерна, его плюсы и минусы,а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Facade_pattern
*/

/*
Facade - структура-фасад, скрывающая за собой в данном случае две сложные структуры.

Цель такого сокрытия - обеспечить взаимодействие с несколькими сложными структурами как с единым целым. Выделяются
некоторые важные действия, которые могут выполнять вместе скрытые за фасадом структуры, и описываются в виде метода
для структуры-фасада.

Преимущество - возможность создания интерфейса для простого взаимодействия со сложными структурами, уменьшение
дублирования кода. Недостатки - меньше контроля над сокрытыми структурами при использовании фасада.

Примером может служить, например, API, скрывающий за собой микросервисную архитектуру. Клиент не взаимодействует с
каждым сервисом напрямую, а вместо этого делает запрос в API, играющий в данном случае роль фасада. API делает запросы
в нужные сервисы и возвращает ответ клиенту.
*/
type Facade struct {
	subsystem1 *ComplexSubsystem1
	subsystem2 *ComplexSubsystem2
}

func NewFacade(subsystem1 *ComplexSubsystem1, subsystem2 *ComplexSubsystem2) *Facade {
	return &Facade{subsystem1: subsystem1, subsystem2: subsystem2}
}

func (f *Facade) SomeComplexAction() {
	f.subsystem2.SomeAction2()
	f.subsystem1.SomeAction1()
	f.subsystem2.SomeAction1()
	f.subsystem1.SomeAction2()
}

type ComplexSubsystem1 struct{}

func (c *ComplexSubsystem1) SomeAction1() {}

func (c *ComplexSubsystem1) SomeAction2() {}

type ComplexSubsystem2 struct{}

func (c *ComplexSubsystem2) SomeAction1() {}

func (c *ComplexSubsystem2) SomeAction2() {}
