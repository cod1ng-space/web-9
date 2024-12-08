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
func (h *Handlers) GetCount(c echo.Context) error {
	msg, err := h.dbProvider.SelectCount()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else {
		return c.String(http.StatusOK, msg)
	}
}
func (h *Handlers) PostCount(c echo.Context) error {
	input := struct {
		Msg *int `json:"count"`
	}{}
	err := c.Bind(&input)
	if input.Msg == nil {
		return c.String(http.StatusBadRequest, "Отсутствует поле 'count'!")
	} else if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	} else {
		err = h.dbProvider.UpdateCount(*input.Msg)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		} else {
			return c.String(http.StatusAccepted, "")
		}
	}
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectCount() (string, error) {
	var msg string

	row := dp.db.QueryRow("SELECT num FROM counter")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}
func (dp *DatabaseProvider) UpdateCount(msg int) error {
	_, err := dp.db.Exec("UPDATE counter SET num = num + $1", msg)
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
	e.GET("/get", h.GetCount)
	e.POST("/post", h.PostCount)

	fmt.Println("Сервер запущен")
	// Запускаем веб-сервер на указанном адресе
	e.Logger.Fatal(e.Start(":8083"))
}
