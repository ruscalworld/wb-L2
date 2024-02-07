package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

/*
=== HTTP server ===

Реализовать HTTP сервер для работы с календарем. В рамках задания необходимо работать строго со стандартной HTTP библиотекой.
В рамках задания необходимо:
	1. Реализовать вспомогательные функции для сериализации объектов доменной области в JSON.
	2. Реализовать вспомогательные функции для парсинга и валидации параметров методов /create_event и /update_event.
	3. Реализовать HTTP обработчики для каждого из методов API, используя вспомогательные функции и объекты доменной области.
	4. Реализовать middleware для логирования запросов
Методы API: POST /create_event POST /update_event POST /delete_event GET /events_for_day GET /events_for_week GET /events_for_month
Параметры передаются в виде www-url-form-encoded (т.е. обычные user_id=3&date=2019-09-09).
В GET методах параметры передаются через queryString, в POST через тело запроса.
В результате каждого запроса должен возвращаться JSON документ содержащий либо {"result": "..."} в случае успешного выполнения метода,
либо {"error": "..."} в случае ошибки бизнес-логики.

В рамках задачи необходимо:
	1. Реализовать все методы.
	2. Бизнес логика НЕ должна зависеть от кода HTTP сервера.
	3. В случае ошибки бизнес-логики сервер должен возвращать HTTP 503. В случае ошибки входных данных (невалидный int например) сервер должен возвращать HTTP 400. В случае остальных ошибок сервер должен возвращать HTTP 500. Web-сервер должен запускаться на порту указанном в конфиге и выводить в лог каждый обработанный запрос.
	4. Код должен проходить проверки go vet и golint.
*/

// ----- БИЗНЕС-ЛОГИКА -----

// Event представляет собственно события, с которыми мы будем работать
type Event struct {
	Name string
	Date time.Time
}

// Calendar управляет событиями. Будем хранить данные в простой мапе.
type Calendar struct {
	events map[int64]*Event
	lock   *sync.RWMutex
	lastId int64
}

func NewCalendar() *Calendar {
	return &Calendar{
		events: make(map[int64]*Event),
		lock:   &sync.RWMutex{},
	}
}

// CreateEvent создаёт новое событие
func (c *Calendar) CreateEvent(event *Event) int64 {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Каждый новый ID будет на 1 больше предыдущего
	c.lastId++

	// Сохраняем событие в мапу
	c.events[c.lastId] = event
	return c.lastId
}

// UpdateEvent обновляет событие по идентификатору
func (c *Calendar) UpdateEvent(id int64, event *Event) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Не даём создавать новые события
	if _, ok := c.events[id]; !ok {
		return false
	}

	// Обновляем событие в мапе
	c.events[id] = event
	return true
}

// DeleteEvent удаляет событие по ID
func (c *Calendar) DeleteEvent(id int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Просто удаляем событие из мапы по ключу
	delete(c.events, id)
}

// GetEvents возвращает события в заданном диапазоне
func (c *Calendar) GetEvents(since, till time.Time) map[int64]*Event {
	c.lock.RLock()
	defer c.lock.RUnlock()

	result := make(map[int64]*Event)

	// Проходим по всем событиям
	for id, event := range c.events {

		// Отсеиваем неподходящие нам события
		if event.Date.Before(since) || event.Date.After(till) {
			continue
		}

		// Подходящие события записываем в результат
		result[id] = event
	}

	return result
}

// ----- HTTP СЕРВЕР -----

// Формат для дат, который будет использоваться как для парсинга, так и для форматирования
const dateFormat = "2006-01-02"

// Response - обёртка для ответов сервера
type Response struct {
	Result any `json:"result"`
}

// ErrorResponse - обёртка для обработанных ошибок
type ErrorResponse struct {
	code    int
	Message string `json:"error"`
}

func NewErrorResponse(code int, message string) *ErrorResponse {
	return &ErrorResponse{code: code, Message: message}
}

func (e *ErrorResponse) Error() string {
	return e.Message
}

// write записывает в http.ResponseWriter данную ошибку
func (e *ErrorResponse) write(w http.ResponseWriter) {
	writeResponse(w, e, e.code)
}

// IdRequest - запрос, содержащий только идентификатор события
type IdRequest struct {
	ID int64 `json:"id"`
}

// Parse загружает значения из данного url.Values в структуру
func (r *IdRequest) Parse(values url.Values) error {
	var err error

	// Преобразуем параметр id в число
	r.ID, err = strconv.ParseInt(values.Get("id"), 10, 64)
	if err != nil {
		return NewErrorResponse(http.StatusBadRequest, err.Error())
	}

	return nil
}

// EventRequest - запрос, содержащий информацию о событии
type EventRequest struct {
	Name string
	Date time.Time
}

// Parse загружает значения из данного url.Values в структуру
func (e *EventRequest) Parse(values url.Values) error {
	var err error

	// Преобразуем параметр date в time.Time
	e.Date, err = time.Parse(dateFormat, values.Get("date"))
	if err != nil {
		return NewErrorResponse(http.StatusBadRequest, err.Error())
	}

	// Получаем и проверяем параметр name
	e.Name = strings.TrimSpace(values.Get("name"))
	if e.Name == "" {
		return NewErrorResponse(http.StatusBadRequest, "name cannot be empty")
	}

	return nil
}

// EventRequestWithId - запрос, содержащий как информацию о событии, так и идентификатор
type EventRequestWithId struct {
	IdRequest
	EventRequest
}

// Parse загружает значения из данного url.Values в структуру
func (e *EventRequestWithId) Parse(values url.Values) error {
	// Загружаем данные в IdRequest
	err := e.IdRequest.Parse(values)
	if err != nil {
		return err
	}

	// Загружаем данные в EventRequest
	return e.EventRequest.Parse(values)
}

// EventResponse описывает структуру ответа сервера, содержащего информацию о событии
type EventResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	Date string `json:"date"`
}

// handler - функция, обрабатывающая запрос и возвращающая некоторый ответ или ошибку
type handler func(r *http.Request) (any, error)

// CalendarServer содержит методы, реализующие функционал API
type CalendarServer struct {
	Calendar *Calendar
}

func NewCalendarServer(calendar *Calendar) *CalendarServer {
	return &CalendarServer{Calendar: calendar}
}

// Mux создаёт http.ServeMux со всеми обработчиками методов API
func (s *CalendarServer) Mux() *http.ServeMux {
	mux := new(http.ServeMux)

	mux.Handle("/create_event", makeHandler(http.MethodPost, s.CreateEvent))
	mux.Handle("/update_event", makeHandler(http.MethodPost, s.UpdateEvent))
	mux.Handle("/delete_event", makeHandler(http.MethodPost, s.DeleteEvent))
	mux.Handle("/events_for_day", makeHandler(http.MethodGet, s.EventsForDay))
	mux.Handle("/events_for_week", makeHandler(http.MethodGet, s.EventsForWeek))
	mux.Handle("/events_for_month", makeHandler(http.MethodGet, s.EventsForMonth))

	return mux
}

// CreateEvent - метод API для создания событий
func (s *CalendarServer) CreateEvent(r *http.Request) (any, error) {
	// Парсим запрос
	request := new(EventRequest)
	err := request.Parse(r.PostForm)
	if err != nil {
		return nil, err
	}

	// Перекладываем параметры запроса в бизнес-логику
	id := s.Calendar.CreateEvent(&Event{
		Name: request.Name,
		Date: request.Date,
	})

	// Формируем ответ
	return &EventResponse{
		ID:   id,
		Name: request.Name,
		Date: request.Date.Format(dateFormat),
	}, nil
}

// UpdateEvent - метод API для обновления событий
func (s *CalendarServer) UpdateEvent(r *http.Request) (any, error) {
	// Парсим запрос
	request := new(EventRequestWithId)
	err := request.Parse(r.PostForm)
	if err != nil {
		return nil, err
	}

	// Перекладываем параметры запроса в бизнес-логику
	ok := s.Calendar.UpdateEvent(request.ID, &Event{
		Name: request.Name,
		Date: request.Date,
	})

	// Возвращаем ошибку, если не удалось обновить событие
	if !ok {
		return nil, NewErrorResponse(http.StatusNotFound, "event not found")
	}

	// Формируем ответ
	return &EventResponse{
		ID:   request.ID,
		Name: request.Name,
		Date: request.Date.Format(dateFormat),
	}, nil
}

// DeleteEvent - метод API для удаления событий
func (s *CalendarServer) DeleteEvent(r *http.Request) (any, error) {
	// Парсим запрос
	request := new(IdRequest)
	err := request.Parse(r.PostForm)
	if err != nil {
		return nil, err
	}

	// Перекладываем параметры запроса в бизнес-логику
	s.Calendar.DeleteEvent(request.ID)
	return nil, nil
}

// EventsForDay - метод API для получения событий для текущего дня
func (s *CalendarServer) EventsForDay(_ *http.Request) (any, error) {
	now := time.Now()

	return s.eventsFor(
		// Начало текущего дня
		time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),

		// Конец текущего дня
		time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()),
	)
}

// EventsForWeek - метод API для получения событий для текущей недели
func (s *CalendarServer) EventsForWeek(_ *http.Request) (any, error) {
	now := time.Now()

	// Номер первого дня текущей недели
	start := now.Day() - int(now.Weekday())

	return s.eventsFor(
		// Начало текущей недели
		time.Date(now.Year(), now.Month(), start, 0, 0, 0, 0, now.Location()),

		// Конец текущей недели
		time.Date(now.Year(), now.Month(), start+7, 23, 59, 59, 0, now.Location()),
	)
}

// EventsForMonth - метод API для получения событий для текущего месяца
func (s *CalendarServer) EventsForMonth(_ *http.Request) (any, error) {
	now := time.Now()

	return s.eventsFor(
		// Начало текущего месяца
		time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()),

		// Конец текущего месяца: текущий месяц + 1 и нулевой день дают последний день текущего месяца
		time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, now.Location()),
	)
}

// eventsFor получает события в заданном диапазоне и формирует ответ
func (s *CalendarServer) eventsFor(since, till time.Time) (any, error) {
	// Получаем события
	events := s.Calendar.GetEvents(since, till)
	result := make([]*EventResponse, len(events))

	// Поскольку далее мы будем обходить мапу, у нас не будет ничего, что можно превратить в индекс элемента в слайсе.
	// Поэтому, сделаем вспомогательную переменную.
	i := 0

	for id, event := range events {
		// Преобразуем структуру события в структуру ответа
		result[i] = &EventResponse{
			ID:   id,
			Name: event.Name,
			Date: event.Date.Format(dateFormat),
		}

		i++
	}

	return result, nil
}

// makeHandler оборачивает функцию-обработчик, преобразуя её возвращаемые значения в ответ HTTP
func makeHandler(method string, handler handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что метод соответствует ожидаемому для данного обработчика
		if r.Method != method {
			NewErrorResponse(http.StatusMethodNotAllowed, "invalid request method").write(w)
			return
		}

		// Для POST запросов парсим тело
		if r.Method == http.MethodPost {
			err := r.ParseForm()
			if err != nil {
				NewErrorResponse(http.StatusBadRequest, err.Error()).write(w)
				return
			}
		}

		// Вызываем сам обработчик
		result, err := handler(r)
		if err != nil {
			var errorResponse *ErrorResponse

			if !errors.As(err, &errorResponse) {
				// Если обработчик вернул неизвестную ошибку, то делаем свою ошибку с кодом 503
				errorResponse = &ErrorResponse{
					code:    http.StatusServiceUnavailable,
					Message: "service error",
				}

				// Заодно логируем эту ошибку
				log.Println("service error:", err)
			}

			// Отправляем ответ
			errorResponse.write(w)
			return
		}

		status := http.StatusOK

		// Если обработчик вернул пустой ответ, то меняем код на 204
		if result == nil {
			status = http.StatusNoContent
		}

		// Наконец отправляем ответ
		writeResponse(w, &Response{result}, status)
	})
}

// writeResponse преобразует ответ в JSON и отправляет его клиенту
func writeResponse(w http.ResponseWriter, data any, status int) {
	// Если код ответа - 204, то просто отправляем этот код. Тело ответа в таком случае отправлять не следует.
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}

	// Сообщаем, что наш ответ будет в формате JSON
	w.Header().Set("Content-Type", "application/json")

	// Отправляем код ответа
	w.WriteHeader(status)

	// Преобразуем ответ в JSON и отправляем его
	encoder := json.NewEncoder(w)
	err := encoder.Encode(data)
	if err != nil {
		log.Println("error encoding response:", err)
		return
	}
}

// Простейший middleware для логирования запросов
func logger(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Засекаем время начала обработки запроса
		start := time.Now()

		// Обрабатываем запрос
		handler.ServeHTTP(w, r)

		// Выводим собранную информацию
		log.Println(r.Method, r.URL.Path, "-", time.Since(start))
	}
}

func main() {
	// Элементарный конфиг через переменные окружения
	bindAddr, ok := os.LookupEnv("BIND_ADDR")
	if !ok {
		log.Fatalln("BIND_ADDR is empty")
		return
	}

	// Инициализируем всё, что нам нужно
	calendar := NewCalendar()
	server := NewCalendarServer(calendar)

	// Запускаем HTTP сервер, используя наш логгер и ServeMux, возвращённый структурой сервера
	err := http.ListenAndServe(bindAddr, logger(server.Mux()))
	if err != nil {
		log.Fatalln(err)
		return
	}
}
