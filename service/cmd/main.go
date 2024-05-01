package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/VoC925/go_final_project/service/internal/api"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	var (
		signalCh = make(chan os.Signal, 1) // сигнал ОС
		quitCh   = make(chan struct{})     // канал выхода из программы
	)
	// регистрация сигнала ОС
	signal.Notify(signalCh, os.Interrupt)

	port := os.Getenv("TODO_PORT")
	app, err := api.NewApp(port)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"port": port,
		}).Error(errors.Wrap(err, "app"))
		return
	}

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
