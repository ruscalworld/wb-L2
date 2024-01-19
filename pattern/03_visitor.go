package pattern

/*
	Реализовать паттерн «посетитель».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Visitor_pattern

Паттерн позволяет реализовывать функции, использующие ту или иную структуру, не внося изменения в саму структуру.
Реализуется структура-посетитель, которая реализует некоторый интерфейс. Структура-посетитель имеет метод, принимающий
указатель на структуру, с которой происходит взаимодействие. Когда нужно выполнить то или иное действие,
экземпляр структуры-посетителя передаётся в метод целевой структуры.

Преимущества - возможность реализовать любое количество возможных операций над той или иной структурой без внесения
изменений в неё.

Пример использования: есть структура, которую нужно обновлять по разным алгоритмам. В таком случае для неё реализуются
посетители, каждый из которых выполняет определённое действие над структурой.
*/

// Visitor описывает общий тип для посетителей структуры Target.
type Visitor interface {
	Visit(*Target)
}

// Target - структура, над которой будем выполнять какие-либо действия.
type Target struct {
	A int
}

// Accept принимает посетителя и передаёт ему ссылку на текущий экземпляр структуры.
func (t *Target) Accept(visitor Visitor) {
	visitor.Visit(t)
}

// Visitor1 - первый вариант посетителя
type Visitor1 struct{}

func (v Visitor1) Visit(target *Target) {
	target.A = 1
}

// Visitor2 - первый вариант посетителя
type Visitor2 struct{}

func (v Visitor2) Visit(target *Target) {
	target.A = 2
}
