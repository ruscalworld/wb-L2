package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

/*
=== Утилита sort ===

Отсортировать строки (man sort)
Основное

Поддержать ключи

-k — указание колонки для сортировки
-n — сортировать по числовому значению
-r — сортировать в обратном порядке
-u — не выводить повторяющиеся строки

Дополнительное

Поддержать ключи

-M — сортировать по названию месяца
-b — игнорировать хвостовые пробелы
-c — проверять отсортированы ли данные
-h — сортировать по числовому значению с учётом суффиксов

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

var (
	// Индекс столбца, -1 для выбора строки целиком вне зависимости от кол-ва столбцов
	columnIndex = flag.Int("k", -1, "column index")

	// Использовать числовое значение вместо строкового
	arithmeticValue = flag.Bool("n", false, "use arithmetic value")

	// Сортировать в обратном порядке
	reverseOrder = flag.Bool("r", false, "sort in reverse order")

	// Оставить строки только с уникальными ключами
	//
	// Из man sort:
	// -u, --unique
	//         Unique keys.  Suppress all lines that have a key that is equal to an already processed one.
	// Реализовано именно такое поведение - исключение дубликатов по ключу, а не по содержанию оригинальной строки.
	unique = flag.Bool("u", false, "keep only unique lines")
)

// FileHolder хранит информацию о данных, которые были получены из файла, а также настройки для его сортировки.
// Реализует sort.Interface, за счёт чего экземпляр этой структуры может быть передан в sort.Sort.
type FileHolder struct {
	lines []string

	columnIndex     int
	arithmeticValue bool
	reverseOrder    bool
	unique          bool
}

// NewFileHolder создаёт новый пустой FileHolder. Для работы требуется далее вызвать метод FileHolder.ReadLines.
func NewFileHolder() *FileHolder {
	return &FileHolder{
		columnIndex:     *columnIndex,
		arithmeticValue: *arithmeticValue,
		reverseOrder:    *reverseOrder,
		unique:          *unique,
	}
}

// ReadLines считывает построчно данные из переданного reader и сохраняет строки в FileHolder.
func (h *FileHolder) ReadLines(reader io.Reader) {
	h.lines = make([]string, 0)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		h.lines = append(h.lines, scanner.Text())
	}
}

// WriteOutput записывает в переданный writer отсортированные данные, находящиеся в FileHolder.
func (h *FileHolder) WriteOutput(writer io.Writer) (n int, err error) {
	uniqueKeys := make(map[any]bool, h.Len())

	for _, line := range h.lines {
		k := h.Key(line)
		// Проверяем, не записывали ли мы до этого строку с таким же ключом. Если передан флаг, требующий уникальности
		// каждой строки в выходных данных, то пропускаем текущую строку.
		if _, ok := uniqueKeys[k]; ok && h.unique {
			continue
		}

		// Записываем текущую строку
		m, err := writer.Write([]byte(line + "\n"))
		n += m
		if err != nil {
			return n, err
		}

		uniqueKeys[k] = true
	}

	return
}

func (h *FileHolder) Len() int {
	return len(h.lines)
}

// Less сравнивает строки с индексами i и j с учётом заданных параметров сортировки. Возвращается true, если элемент с
// индексом i должен стоять перед элементом с индексом j.
func (h *FileHolder) Less(i, j int) bool {
	// Получаем ключи для каждой из строк, которые мы будем непосредственно сравнивать
	a := h.Key(h.lines[i])
	b := h.Key(h.lines[j])

	// Если требуется отсортировать в обратном порядке, достаточно поменять ключи местами
	if h.reverseOrder {
		a, b = b, a
	}

	// Приводим значения ключей к конкретным типам, чтобы иметь возможность сравнить их через операцию "<"
	switch a.(type) {
	case string:
		return a.(string) < b.(string)
	case float64:
		return a.(float64) < b.(float64)
	default:
		panic("unsupported types")
	}
}

// Swap меняет местами элементы на позициях i и j
func (h *FileHolder) Swap(i, j int) {
	h.lines[i], h.lines[j] = h.lines[j], h.lines[i]
}

// Key получает ключ, который будет использоваться непосредственно для сравнения элементов при сортировке, для строки
// input. Учитывает возможность разбиения на столбцы и использования в качестве ключа числового значения вместо строкового.
func (h *FileHolder) Key(input string) any {
	// Если задан определённый индекс столбца, по которому нужно сортировать, разбиваем входную строку на столбцы и
	// выбираем столбец с нужным индексом. Если индекс находится вне границ массива, игнорируем разбиение по столбцам и
	// работаем со всей строкой целиком.
	if h.columnIndex >= 0 {
		columns := strings.Split(input, " ")
		if h.columnIndex < len(columns) {
			input = columns[h.columnIndex]
		}
	}

	// Если требуется использовать числовое значение вместо строкового, то преобразуем входную строку в число. Если
	// преобразование не удаётся, используем в качестве ключа 0.
	if h.arithmeticValue {
		float, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return 0.0
		}

		return float
	}

	return input
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("filename not provided")
		os.Exit(1)
		return
	}

	// Открываем файл с переданным через аргументы названием
	file, err := os.Open(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
		return
	}

	// Закрываем файл при выходе из программы
	defer file.Close()

	// Инициализируем FileHolder
	s := NewFileHolder()
	s.ReadLines(file)

	// Осуществляем сортировку
	sort.Sort(s)

	// По умолчанию выводим в stdout.
	// Если передано два аргумента, то считаем последний из них названием выходного файла.
	out := os.Stdout
	if len(args) >= 2 {
		file, err := os.Create(args[1])
		if err != nil {
			fmt.Println("unable to open output file:", err)
		} else {
			out = file
			defer file.Close()
		}
	}

	// Записываем выходные данные
	_, err = s.WriteOutput(out)
	if err != nil {
		fmt.Println(err)
		os.Exit(3)
		return
	}
}
