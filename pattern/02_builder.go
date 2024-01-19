package pattern

/*
	Реализовать паттерн «строитель».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Builder_pattern
*/

/*
Паттерн Builder применяется, когда нужно упростить процесс создания сложной структуры. Реализация структуры-строителя
позволяет постепенно указывать различные свойства некоторой сложной структуры, и после этого вызвать метод, который
непосредственно инициализирует экземпляр данной сложной структуры.

Преимущества - позволяет разделить на несколько этапов создание сложной структуры. Недостатки - сложность поддержки:
при добавлении новых полей в структуру их также необходимо добавлять в структуру-строитель.

Пример применения - любая сложная структура (имеющая большое количество полей), взаимодействие с которой можно
упростить, реализовав структуру-строитель.
*/

// ComplexStruct - некоторая сложная структура.
type ComplexStruct struct {
	A, B, C, D, E int
}

// ComplexStructBuilder - структура-строитель, которая создаёт экземпляр структуры ComplexStruct.
type ComplexStructBuilder struct {
	a, b, c, d, e int
}

func (c *ComplexStructBuilder) A(A int) {
	c.a = A
}

func (c *ComplexStructBuilder) B(B int) {
	c.b = B
}

func (c *ComplexStructBuilder) C(C int) {
	c.c = C
}

func (c *ComplexStructBuilder) D(D int) {
	c.d = D
}

func (c *ComplexStructBuilder) E(E int) {
	c.e = E
}

// Build создаёт экземпляр ComplexStruct, используя полученные ранее данные.
func (c *ComplexStructBuilder) Build() *ComplexStruct {
	return &ComplexStruct{
		A: c.a,
		B: c.b,
		C: c.c,
		D: c.d,
		E: c.e,
	}
}
