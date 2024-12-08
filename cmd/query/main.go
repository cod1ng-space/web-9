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
func (h *Handlers) GetQuery(c echo.Context) error {

	msg, err := h.dbProvider.SelectQuery()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else if msg == "" {
		return c.String(http.StatusBadRequest, "Неинициализированное значение в базе данных!")
	} else {
		return c.String(http.StatusOK, "Hello, "+msg+"!")
	}
}
func (h *Handlers) PostQuery(c echo.Context) error {

	input := struct {
		Msg *string `json:"name"`
	}{}

	err := c.Bind(&input)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	} else if input.Msg == nil {
		return c.String(http.StatusBadRequest, "Отсутствует поле 'name'!")
	} else if *input.Msg == "" {
		return c.String(http.StatusBadRequest, "Пустая строка!")
	} else {
		err = h.dbProvider.UpdateQuery(*input.Msg)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		} else {
			return c.String(http.StatusAccepted, "")
		}
	}
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectQuery() (string, error) {
	var msg string

	row := dp.db.QueryRow("SELECT name FROM query")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}
func (dp *DatabaseProvider) UpdateQuery(msg string) error {
	_, err := dp.db.Exec("UPDATE query SET name = $1", msg)
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
	e.GET("/get", h.GetQuery)
	e.POST("/post", h.PostQuery)

	fmt.Println("Сервер запущен")
	// Запускаем веб-сервер на указанном адресе
	e.Logger.Fatal(e.Start(":8082"))
}
