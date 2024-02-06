package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"strings"
	"sync"
)

/*
=== Утилита wget ===

Реализовать утилиту wget с возможностью скачивать сайты целиком

Программа должна проходить все тесты. Код должен проходить проверки go vet и golint.
*/

var maxDepth = flag.Int("d", 1, "depth")

// Функция download скачивает страницу по адресу address в директорию dst. К пути dst прибавляется путь к странице
// внутри сайта.
func download(address string, dst string, downloaded *sync.Map, depth int) error {
	// Запоминаем, что мы скачали эту страницу
	downloaded.Store(address, true)
	fmt.Println("Downloading", address)

	// Скачиваем страницу
	response, err := http.Get(address)
	if err != nil {
		return err
	}

	// Парсим адрес страницы
	u, err := url.Parse(address)
	if err != nil {
		return err
	}

	// Приводим URL к такому виду, что мы сможем получить путь к файлу, соответствующему ему в локальной файловой системе
	normalize(u)

	// Формируем путь к файлу, в который будем сохранять страницу
	root := path.Join(dst, u.Host)
	targetPath := path.Join(root, u.Path)

	// Создаём нужные директории
	err = os.MkdirAll(path.Dir(targetPath), 0750)
	if err != nil {
		return err
	}

	// Создаём сам файл
	file, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	// Читаем всё содержимое тела ответа
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// Сохраняем тело ответа в файл
	_, err = file.Write(body)
	if err != nil {
		return err
	}

	// Закрываем файл
	err = file.Close()
	if err != nil {
		return err
	}

	// Если мы достигли максимальной глубины, то дальше ничего делать не нужно
	if depth == *maxDepth {
		return nil
	}

	// Ищем в HTML файлах ссылки на дополнительные ресурсы
	if strings.Split(response.Header.Get("Content-Type"), ";")[0] == "text/html" {
		// Достаём ссылки на эти ресурсы и скачиваем их при необходимости
		resources := ExtractResources(body)
		processResources(address, dst, downloaded, resources, depth)
	}

	return nil
}

func normalize(u *url.URL) {
	// Мы не можем сохранить файл в путь "/", поэтому считаем, что главная страница - index.html
	if u.Path == "/" || u.Path == "" {
		u.Path = "/index.html"
	}

	// Если для страницы не указано расширение, то считаем, что подразумевалось расширение .html
	if path.Ext(u.Path) == "" {
		u.Path = u.Path + ".html"
	}
}

func processResources(root string, dst string, downloaded *sync.Map, resources []string, depth int) {
	for _, resource := range resources {
		// Получаем путь к ресурсу относительно запрошенной страницы нашего сайта
		resourceUrl, err := getAbsolutePath(root, resource)
		if err != nil {
			fmt.Println("malformed resource url:", resource)
			continue
		}

		// Проверяем, нужно ли загружать этот ресурс
		if !shouldDownload(root, resourceUrl) {
			continue
		}

		// Если он уже был загружен ранее, пропускаем его
		if _, ok := downloaded.Load(resourceUrl); ok {
			continue
		}

		// Собственно загружаем ресурс
		err = download(resourceUrl, dst, downloaded, depth+1)
		if err != nil {
			fmt.Println(err)
			continue
		}
	}
}

func getAbsolutePath(root, resourceUrl string) (string, error) {
	u, err := url.Parse(root)
	if err != nil {
		return "", nil
	}

	if strings.HasPrefix(resourceUrl, "/") {

		// Если мы получили адрес, начинающийся с / (например, /main.css), то считаем, что это адрес относительно
		// корня нашего сайта. В таком случае просто заменяем путь в URL корня нашего сайта на полученный.

		u.Path = resourceUrl
		return u.String(), nil
	} else if strings.Contains(resourceUrl, "://") {

		// Если мы получили адрес, содержащий ://, то значит, что это сам по себе абсолютный адрес.
		// Тогда достаточно просто вернуть этот адрес без дополнительных манипуляций.

		return resourceUrl, nil
	}

	// Во всех остальных случаях добавляем адрес ресурса в конец изначального адреса.
	u.Path = path.Join(u.Path, resourceUrl)
	return u.String(), nil
}

func shouldDownload(root string, resource string) bool {
	// То, что начинается с "data:" не является адресом
	if strings.HasPrefix(resource, "data:") {
		return false
	}

	a, err := url.Parse(root)
	if err != nil {
		return false
	}

	b, err := url.Parse(resource)
	if err != nil {
		return false
	}

	// Скачиваем ресурсы только с одного домена
	return a.Host == b.Host
}

// Паттерн, который позволяет извлечь значения атрибутов href и src
var resourceUrlPattern = regexp.MustCompile(`<[^<]+(href|src)=["']([^'"]+)["'][^>]*>`)

// ExtractResources принимает на вход текст в формате html и возвращает найденные в нём ссылки
func ExtractResources(html []byte) []string {
	// Пропускаем наш html через паттерн
	matches := resourceUrlPattern.FindAllSubmatch(html, -1)
	result := make([]string, len(matches))

	// Из каждой найденной подстроки извлекаем нужный нам фрагмент
	for i, m := range matches {
		result[i] = string(m[2])
	}

	return result
}

func main() {
	flag.Parse()
	u := flag.Arg(0)

	err := download(u, ".", &sync.Map{}, 1)
	if err != nil {
		fmt.Println(err)
		return
	}
}
