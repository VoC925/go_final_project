package main

import (
	"os"
	"os/signal"

	"github.com/VoC925/go_final_project/internal/api"
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
		if err := app.Start(); err != nil {
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
}

func init() {
	// загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logrus.Fatal("Error loading .env file")
	}

	// установка уровня логирования debug в случае установки переменной окружения
	if os.Getenv("IS_DEBUG_LOG_LEVEL") == "true" {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.WithFields(logrus.Fields{
		"TODO_PORT":          os.Getenv("TODO_PORT"),
		"TODO_DBFILE":        os.Getenv("TODO_DBFILE"),
		"IS_DEBUG_LOG_LEVEL": os.Getenv("IS_DEBUG_LOG_LEVEL"),
		"TODO_PASSWORD":      os.Getenv("TODO_PASSWORD"),
		"JWT_SECRET":         os.Getenv("JWT_SECRET"),
	}).Debug("environment variable")
}
