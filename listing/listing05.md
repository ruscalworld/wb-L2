Что выведет программа? Объяснить вывод программы.

```go
package main

type customError struct {
	msg string
}

func (e *customError) Error() string {
	return e.msg
}

func test() *customError {
	{
		// do something
	}
	return nil
}

func main() {
	var err error
	err = test()
	if err != nil {
		println("error")
		return
	}
	println("ok")
}
```

Ответ:
```
error
```

Чтобы интерфейс был равен `nil`, пустыми должны быть как тип значения в интерфейсе, так и само значение. Поскольку
функция `test` имеет возвращаемое значение типа `*customError`, то при возвращении значения `nil` фактически будет
возвращено `(*customError)(nil)`, хранящее информацию о типе. Это значение не равно `nil`, поэтому будет выведено
`error`.

По-хорошему функция `test` должна возвращать значение типа `error`. Поскольку `error` является интерфейсом, при возврате
`nil` конкретный тип будет неизвестен, и в итоге будет возвращён "истинный" `nil` без указания конкретного типа.
