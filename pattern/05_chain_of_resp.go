package pattern

import "fmt"

/*
	Реализовать паттерн «цепочка вызовов».
Объяснить применимость паттерна, его плюсы и минусы, а также реальные примеры использования данного примера на практике.
	https://en.wikipedia.org/wiki/Chain-of-responsibility_pattern
*/

/*
Паттерн позволяет разделить выполнение некоторой последовательности действий на несколько различных этапов, за
выполнение каждого из которых будет отвечать некоторый обработчик. Обработчик может либо выполнить какое-либо действие,
либо передать ответственность следующему обработчику, если таковой имеется.

Преимущества - возможность разделить код на отдельные компоненты, каждый из которых будет ответственным за выполнение
одного определённого действия. Недостатки - ответствие гарантии того, что какой-либо из обработчиков в итоге сработает.

Пример - логика работы middleware в подавляющем большинстве фреймворков для разработки веб-приложений. Каждый middleware
может либо приостановить выполнение запроса, сразу вернув ответ, либо передать ответственность за дальнейшую обработку
запроса следующему обработчику.
*/

// ChainTarget - пример некоторой сущности, которая будет передаваться в цепочку из обработчиков.
type ChainTarget struct {
	A int
}

// ChainHandler - общий тип для всех элементов цепочки ответственности.
type ChainHandler interface {
	// GetNext возвращает следующий обработчик в цепочке ответственности.
	GetNext() ChainHandler

	// Handle выполняет логику обработки для текущего элемента цепочки.
	Handle(target *ChainTarget)
}

type FirstHandler struct{}

func (h *FirstHandler) GetNext() ChainHandler {
	return &SecondHandler{}
}

func (h *FirstHandler) Handle(target *ChainTarget) {
	if target.A == 1 {
		fmt.Println("First handler")
		return
	}

	next := h.GetNext()
	if next != nil {
		next.Handle(target)
	}
}

type SecondHandler struct{}

func (h *SecondHandler) GetNext() ChainHandler {
	return nil
}

func (h *SecondHandler) Handle(target *ChainTarget) {
	if target.A == 2 {
		fmt.Println("Second handler")
		return
	}

	next := h.GetNext()
	if next != nil {
		next.Handle(target)
	}
}
