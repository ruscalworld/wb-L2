package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

/*
=== Утилита grep ===

Реализовать утилиту фильтрации (man grep)

Поддержать флаги:
-A - "after" печатать +N строк после совпадения
-B - "before" печатать +N строк до совпадения
-C - "context" (A+B) печатать ±N строк вокруг совпадения
-c - "count" (количество строк)
-i - "ignore-case" (игнорировать регистр)
-v - "invert" (вместо совпадения, исключать)
-F - "fixed", точное совпадение со строкой, не паттерн
-n - "line num", печатать номер строки

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

var (
	after      = flag.Int("A", 0, "lines after match")
	before     = flag.Int("B", 0, "lines before match")
	context    = flag.Int("C", 0, "lines around match")
	count      = flag.Bool("c", false, "print only count")
	ignoreCase = flag.Bool("i", false, "ignore case")
	invert     = flag.Bool("v", false, "invert predicate")
	fixed      = flag.Bool("F", false, "only exact matches")
	lineNumber = flag.Bool("n", false, "print line number")
)

// LinePredicate - тип для функции, которая принимает строку
type LinePredicate func(string) bool

// Search отвечает за выбор строк, соответствующих условиям поиска
type Search struct {
	// Содержит строки, среди которых мы буем производить поиск
	sourceLines []string

	// Заполняется методом Find и содержит сопоставление "номер строки => подходит ли под условия"
	matches map[int]bool

	// Нужно ли инвертировать результаты поиска?
	invert bool

	// Заполняется методом Find и хранит количество строк, соответствующих условиям поиска
	count int
}

func NewSearch(lines []string) *Search {
	return &Search{
		sourceLines: lines,
		matches:     make(map[int]bool),
		invert:      *invert,
	}
}

// Find осуществляет поиск среди сохранённых в структуре строк, используя функцию predicate.
func (s *Search) Find(predicate LinePredicate) {
	for i, line := range s.sourceLines {
		// Проверяем, соответствует ли строка
		r := predicate(line)

		// Если нужно, инвертируем результат
		if s.invert {
			r = !r
		}

		// Сохраняем результат
		s.matches[i] = r

		// Инкремент количества найденных строк
		if r {
			s.count++
		}
	}
}

// Printer отвечает за вывод результата поиска
type Printer struct {
	// Количество выводимых строк до соответствующей условиям
	linesBefore int

	// Количество выводимых строк после соответствующей условиям
	linesAfter int

	// Выводить номера строк?
	lineNumbers bool

	// Вывести только количество строк?
	onlyCount bool
}

func NewPrinter() *Printer {
	if *context != 0 {
		*before = *context
		*after = *context
	}

	return &Printer{
		linesBefore: *before,
		linesAfter:  *after,
		lineNumbers: *lineNumber,
		onlyCount:   *count,
	}
}

func (p *Printer) Print(search *Search, writer io.Writer) (n int, err error) {
	// Если нужно вывести только количество, то выводим и сразу выходим
	if p.onlyCount {
		return fmt.Fprintf(writer, "%d\n", search.count)
	}

	// Флаг, который отвечает за то, что нужно выводить строку
	shouldPrint := false

	// Переменная для индекса последней найденной строки
	lastMatch := -1

	for i, line := range search.sourceLines {
		// Эту строки (и, возможно, последующие строки) выводим при одном из условий:
		// - после обработки предыдущей строки shouldPrint имеет значение true
		// - текущая строка соответствует условиям поиска
		// - строка на позиции i + linesBefore соответствует условиям поиска
		currentMatches := search.matches[i]
		hasMatchFurther := i+p.linesBefore < len(search.sourceLines) && search.matches[i+p.linesBefore]
		shouldPrint = shouldPrint || currentMatches || hasMatchFurther

		// Если текущая строка соответствует условиям поиска, сохраняем её индекс
		if currentMatches {
			lastMatch = i
		}

		if shouldPrint {
			// Выводим номер строки, если нужно
			if p.lineNumbers {
				// Как в оригинальном grep: если текущая строка соответствует условиям, выводим ":", иначе - "-"
				matchIndicator := "-"
				if currentMatches {
					matchIndicator = ":"
				}

				// Записываем номер строки и индикатор соответствия условиям поиска
				m, err := fmt.Fprintf(writer, "%d%s", i+1, matchIndicator)
				n += m
				if err != nil {
					return n, err
				}
			}

			// Выводим саму строку
			m, err := writer.Write([]byte(line + "\n"))
			n += m
			if err != nil {
				return n, err
			}
		}

		// Если текущая строка является последней из тех, которые нужно вывести после найденной, то отключаем вывод для
		// последующих строк.
		if lastMatch+p.linesAfter <= i {
			shouldPrint = false
		}
	}

	return
}

// ReadInput получает входные данные либо из файла, если задано его название, либо из stdin.
func ReadInput() []string {
	// Используем stdin по умолчанию
	input := os.Stdin

	// Если указано название файла, то открываем его и используем в качестве ввода
	if fileName := flag.Arg(1); fileName != "" {
		file, err := os.Open(fileName)
		if err != nil {
			fmt.Println(err)
			os.Exit(3)
			return nil
		}

		input = file
	}

	return ReadLines(input)
}

// ReadLines извлекает список строк из reader
func ReadLines(reader io.ReadCloser) []string {
	defer reader.Close()

	scanner := bufio.NewScanner(reader)
	lines := make([]string, 0)

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines
}

// MakePredicate создаёт функцию для проверки строк, исходя из параметров программы.
func MakePredicate(pattern string) (LinePredicate, error) {
	switch {
	case *fixed && *ignoreCase:
		// Строгое соответствие всей строки, но без учёта регистра
		return func(s string) bool {
			return strings.ToLower(s) == strings.ToLower(pattern)
		}, nil

	case *fixed && !*ignoreCase:
		// Строгое соответствие всей строки с учётом регистра
		return func(s string) bool {
			return s == pattern
		}, nil

	default:
		// Соответствие части строки шаблону

		// Если нужно игнорировать регистр, то добавляем (?i) в начало шаблона
		if *ignoreCase {
			pattern = "(?i)" + pattern
		}

		p, err := regexp.Compile(pattern)
		if err != nil {
			return nil, err
		}

		return func(s string) bool {
			return p.MatchString(s)
		}, nil
	}
}

func main() {
	flag.Parse()

	// Считаем первый аргумент шаблоном, по которому будем искать
	pattern := flag.Arg(0)
	if pattern == "" {
		_, _ = fmt.Fprintln(os.Stderr, "no pattern provided")
		os.Exit(1)
		return
	}

	// Составляем функцию для проверки строки
	predicate, err := MakePredicate(pattern)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "invalid pattern:", err)
		os.Exit(2)
		return
	}

	// Считываем входные данные и осуществляем поиск по ним
	lines := ReadInput()
	search := NewSearch(lines)
	search.Find(predicate)

	// Выводим результат
	printer := NewPrinter()
	_, err = printer.Print(search, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(4)
		return
	}
}
