package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

/*
=== Утилита cut ===

Принимает STDIN, разбивает по разделителю (TAB) на колонки, выводит запрошенные

Поддержать флаги:
-f - "fields" - выбрать поля (колонки)
-d - "delimiter" - использовать другой разделитель
-s - "separated" - только строки с разделителем

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

// ColumnSelector хранит информацию о номерах выбранных столбцов
type ColumnSelector struct {
	Ranges []Range
}

func NewColumnSelector(ranges ...Range) ColumnSelector {
	return ColumnSelector{Ranges: ranges}
}

// GetColumns возвращает все номера столбцов в диапазоне [1, max], соответствующие заданным условиям
func (c ColumnSelector) GetColumns(max int) []int {
	values := make(map[int]struct{})

	// Сначала определим, какие у нас вообще выбраны столбцы и поместим их в мапу в качестве ключей
	for _, r := range c.Ranges {
		for _, v := range r.GetEntries(1, max) {
			// Исключим всё, что находится за пределами диапазона
			if v < 1 || v > max {
				continue
			}

			// Запоминаем, что нужно добавить столбец с номером v в результат
			values[v] = struct{}{}
		}
	}

	// Собираем ключи из мапы в слайс
	result := make([]int, 0, len(values))
	for col := range values {
		result = append(result, col)
	}

	// Сортируем результат и возвращаем его
	sort.Ints(result)
	return result
}

// ParseColumnSelector преобразует список столбцов из строкового формата в структуру, с которой далее работает программа
func ParseColumnSelector(input string) (ColumnSelector, error) {
	// Считаем, что через запятую перечислены диапазоны значений (например, "1-2,4" состоит из диапазонов "1-2" и "4").
	// Диапазонами далее будем считать как конкретные значения, так и не более двух значений, разделённых "-".
	parts := strings.Split(input, ",")
	ranges := make([]Range, len(parts))

	// Парсим каждый диапазон и сохраняем его в слайс
	for i, rawRange := range parts {
		r, err := ParseRange(rawRange)
		if err != nil {
			return ColumnSelector{}, fmt.Errorf("error parsing range \"%s\": %s", rawRange, err)
		}

		ranges[i] = r
	}

	return NewColumnSelector(ranges...), nil
}

// Range хранит информацию о конкретном диапазоне
type Range struct {
	// Bounds должен может содержать либо одно, либо два значения:
	// - При длине 1 считается, что диапазон хранит одно конкретное значения, при этом ожидается, что поле
	//   Constraint в структуре границы будет содержать значение 0.
	// - При длине 2 считается, что диапазон хранит некоторое множество чисел, при этом границы могут содержать как
	//   положительное значение поля Constraint, так и положительное значение поля Explicit.
	Bounds []RangeBound
}

func NewRange(bounds ...RangeBound) Range {
	if len(bounds) == 0 || len(bounds) > 2 {
		panic("invalid list of bounds")
	}

	return Range{Bounds: bounds}
}

// GetEntries возвращает список значений, удовлетворяющих данному диапазону. Значения min и max используются только для
// определения границ в случае, если они заданы с помощью ConstraintMin или ConstraintMax.
func (r Range) GetEntries(min, max int) []int {
	switch len(r.Bounds) {
	case 1:
		// Если длина списка границ - 1, то считаем, что у нас должно быть только одно значение
		value := r.Bounds[0]

		// Проверяем, что это значение действительно указано (при отрицательных значениях считаем, что значение
		// не указано)
		if value.Explicit < 0 {
			panic("faced a single bound without a specific value")
			return nil
		}

		// Если всё хорошо, то возвращаем слайс из единственного значения
		return []int{value.Explicit}
	case 2:
		// Если длина списка границ - 1, то считаем, что нужно выбрать все числа из заданного диапазона
		result := make([]int, 0)
		start, end := r.Bounds[0], r.Bounds[1]

		// В цикле перебираем возможные целочисленные значения и добавляем их в результат
		for i := start.GetValue(min, max); i <= end.GetValue(min, max); i++ {
			result = append(result, i)
		}

		// Возвращаем результат
		return result
	default:
		panic("invalid length of bound list")
		return nil
	}
}

// ConstraintType определяет тип неявно заданной границы диапазона
type ConstraintType byte

const (
	// NoConstraint используется, если значение границы задано явно
	NoConstraint ConstraintType = iota

	// ConstraintMin используется для неявно заданной нижней границы диапазона
	ConstraintMin ConstraintType = iota

	// ConstraintMax используется для неявно заданной верхней границы диапазона
	ConstraintMax ConstraintType = iota
)

// RangeBound - граница диапазона
type RangeBound struct {
	// Constraint содержит тип границы - верхняя/нижняя, либо NoConstraint, если указано явное значение в Explicit
	Constraint ConstraintType

	// Explicit содержит числовое значение границы, если она задана явно
	Explicit int
}

func NewSpecificRangeBound(specific int) RangeBound {
	return RangeBound{Explicit: specific, Constraint: NoConstraint}
}

func NewRangeBound(constraint ConstraintType) RangeBound {
	if constraint == NoConstraint {
		panic("attempted to create an empty range")
		return RangeBound{}
	}

	return RangeBound{Explicit: -1, Constraint: constraint}
}

// GetValue возвращает значение границы. Если оно задано неявно, то используется одно из значений min и max в
// зависимости от того, какая это граница - верхняя или нижняя.
func (r RangeBound) GetValue(min, max int) int {
	switch r.Constraint {
	case NoConstraint:
		if r.Explicit < 0 {
			panic("requested a value of an empty bound")
			return 0
		}

		return r.Explicit
	case ConstraintMin:
		return min
	case ConstraintMax:
		return max
	default:
		panic("invalid constraint value")
		return 0
	}
}

// ParseRange преобразует текстовое представление диапазона в структуру Range, которая используется при
// дальнейшей обработке данных
func ParseRange(input string) (Range, error) {
	// Считаем, что границы диапазона разделены "-"
	parts := strings.Split(input, "-")

	switch len(parts) {

	// Если parts состоит из одного элемента, значит, разделителя ("-") в input не было. Считаем, что нам передали
	// границу, заданную явно - с помощью числа.
	case 1:
		raw := parts[0]
		if raw == "" {
			return Range{}, errors.New("empty range value")
		}

		bound, err := parseRangeBound(raw, NoConstraint)
		if err != nil {
			return Range{}, err
		}

		return NewRange(bound), nil

	// Если parts состоит из двух элементов, значит, мы имеем дело со случаем, когда задан сразу целый диапазон.
	// В таком случае парсим левую и правую часть отдельно.
	case 2:
		left, right := parts[0], parts[1]

		leftBound, err := parseRangeBound(left, ConstraintMin)
		if err != nil {
			return Range{}, err
		}

		rightBound, err := parseRangeBound(right, ConstraintMax)
		if err != nil {
			return Range{}, err
		}

		return NewRange(leftBound, rightBound), nil

	// Остальные случаи не поддерживаются.
	default:
		return Range{}, fmt.Errorf("invalid range: %s", input)
	}
}

// parseRangeBound преобразует либо число, либо пустую строку в структуру RangeBound
func parseRangeBound(input string, constraint ConstraintType) (RangeBound, error) {
	// Считаем, что при отсутствии числа слева или справа от дефиса мы имеем дело с неявно заданной нижней или верхней
	// границей соответственно. Слева мы или справа - определяет параметр constraint.
	if input == "" {
		return NewRangeBound(constraint), nil
	}

	// Если строка была непустой, то преобразуем её в число, — получаем явно заданную границу.
	value, err := strconv.Atoi(input)
	if err != nil {
		return RangeBound{}, fmt.Errorf("parsing bound value: %s", err)
	}

	return NewSpecificRangeBound(value), nil
}

// Processor занимается непосредственно обработкой данных, хранит в себе необходимые для этого настройки
type Processor struct {
	fields        ColumnSelector
	delimiter     string
	onlySeparated bool
}

// Process считывает текст из reader и записывает обработанный текст в writer.
func (p *Processor) Process(reader io.Reader, writer io.Writer) (n int, err error) {
	scanner := bufio.NewScanner(reader)

	// Обрабатываем "на лету" каждую строку: считали -> обработали -> вывели -> перешли к следующей строке
	for scanner.Scan() {
		// Получаем считанную строку и пропускаем её через ProcessLine
		line := scanner.Text()
		finalLine, ok := p.ProcessLine(line)

		// Если ProcessLine счёл, что эту строку надо вывести, выводим обработанный её вариант
		if ok {
			m, err := fmt.Fprintln(writer, finalLine)
			n += m
			if err != nil {
				return n, err
			}
		}
	}

	return
}

// ProcessLine обрабатывает одну строку line и возвращает обработанный её вариант, а также флаг, указывающий на то,
// должна ли эта строка присутствовать в выводе
func (p *Processor) ProcessLine(line string) (string, bool) {
	// Если в строке нет разделителей, то обрабатываем её специальным образом
	if !strings.Contains(line, p.delimiter) {
		if p.onlySeparated {
			// Если параметрами задано, что неразделённые строки нужно выбросить, то говорим, что эту строку выводить
			// не нужно
			return "", false
		} else {
			// В противном случае возвращаем строку в чистом виде
			return line, true
		}
	}

	// Разделяем строку с помощью заданного в параметрах разделителя и получаем список номеров столбцов,
	// которые нужно оставить
	allParts := strings.Split(line, p.delimiter)
	requestedColumns := p.fields.GetColumns(len(allParts))
	result := make([]string, len(requestedColumns))

	// Выбираем все столбцы, которые нам нужны, и записываем их в result
	for i, col := range requestedColumns {
		result[i] = allParts[col-1]
	}

	// Соединяем значения из result в одну строку, оставляя прежний разделитель
	return strings.Join(result, p.delimiter), true
}

var (
	fields    = flag.String("f", "", "choose fields")
	delimiter = flag.String("d", "\t", "set specific delimiter")
	separated = flag.Bool("s", false, "keep only lines with delimiter")
)

func main() {
	flag.Parse()

	// Преобразуем "сырые" данные в структуру ColumnSelector, чтобы далее нам было удобнее работать
	selector, err := ParseColumnSelector(*fields)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}

	// Создаём Processor и передаём ему значения флагов
	p := &Processor{
		fields:        selector,
		delimiter:     *delimiter,
		onlySeparated: *separated,
	}

	// Пропускаем весь стандартный ввод через Processor, указывая ему, что вывести всё нужно в стандартный вывод
	_, err = p.Process(os.Stdin, os.Stdout)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
		return
	}
}
