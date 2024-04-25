package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/VoC925/go_final_project/service/internal/api"
	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		signalCh = make(chan os.Signal, 1) // сигнал ОС
		quitCh   = make(chan struct{})     // канал выхода из программы
	)
	// регистрация сигнала ОС
	signal.Notify(signalCh, os.Interrupt)
	// сервис
	route := chi.NewRouter()
	handler := api.NewHandler()
	handler.Register(route)
	port := os.Getenv("TODO_PORT")
	app := api.NewApp(port, route)
	// горутина для запуска сервиса
	go func() {
		fmt.Println("---APP STARTED---")
		if err := app.Start(); err != nil {
			logrus.Error(err)
			close(quitCh)
		}
	}()
	// горутина, слушащая сигнал ОС и завершающая работу сервиса
	go func() {
		<-signalCh
		app.Stop()
		close(quitCh)
	}()

	<-quitCh
	fmt.Println("---APP STOPED---")
}

func init() {
	// загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logrus.Fatal("Error loading .env file")
	}
}
