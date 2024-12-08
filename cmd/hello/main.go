package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgre"
	dbname   = "sandbox"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

// Обработчики HTTP-запросов
func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.String(http.StatusOK, msg)
}
func (h *Handlers) PostHello(c echo.Context) error {
	input := struct {
		Msg *string `json:"msg"`
	}{}

	err := c.Bind(&input)
	if input.Msg == nil {
		return c.String(http.StatusBadRequest, "Отсутствует поле 'msg'!")
	} else if *input.Msg == "" {
		return c.String(http.StatusBadRequest, "Строка пустая!")
	} else if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	err = h.dbProvider.InsertHello(*input.Msg)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusCreated, "")
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string

	// Получаем одно сообщение из таблицы hello, отсортированной в случайном порядке
	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}
func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	e := echo.New()

	// Формирование строки подключения для postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Создание соединения с сервером postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		e.Logger.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер для БД с набором методов
	dp := DatabaseProvider{db: db}

	// Создаем экземпляр структуры с набором обработчиков
	h := Handlers{dbProvider: dp}

	// Регистрируем обработчики
	e.POST("/post", h.PostHello)
	e.GET("/get", h.GetHello)

	fmt.Println("Сервер запущен")
	// Запускаем веб-сервер на указанном адресе
	e.Logger.Fatal(e.Start(":8081"))
}
