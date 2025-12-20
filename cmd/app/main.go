package main

import (
	"html/template"
	"io"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TemplateRegistry подключает html/template к Echo
type TemplateRegistry struct {
	templates *template.Template
}

// Render реализует интерфейс echo.Renderer
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()

	// 1. Мидлвары (логирование и восстановление после паники)
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// 2. Настройка шаблонизатора
	// Парсим все файлы .html из папки web/templates
	templates, err := template.ParseGlob("web/templates/*.html")
	if err != nil {
		log.Fatalf("Ошибка загрузки шаблонов: %v", err)
	}
	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	// 3. Раздача статики (CSS, JS)
	// URL /static/... -> папка web/static
	e.Static("/static", "web/static")

	// 4. Роуты

	// Страница логина
	e.GET("/login", func(c echo.Context) error {
		return c.Render(http.StatusOK, "login.html", nil)
	})

	// Главная страница (защищенная зона - пока имитация)
	e.GET("/", func(c echo.Context) error {
		// В реальном приложении здесь была бы проверка сессии/куки
		// username := c.QueryParam("username") // простая имитация передачи юзера

		data := map[string]interface{}{
			"Title": "DB Explorer Dashboard",
		}
		return c.Render(http.StatusOK, "index.html", data)
	})

	// 5. Запуск сервера
	e.Logger.Fatal(e.Start(":8081"))
}
