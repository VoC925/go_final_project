package api

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/VoC925/go_final_project/service/internal/domain/task"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"

	_ "modernc.org/sqlite"
)

// структура приложения
type App struct {
	// структура сервера
	server *http.Server
	// интерфейс БД
	storage task.Storage
	// канал завершения работы приложения
	quitCh chan struct{}
}

// NewApp создает структуру приложения и привязывает его работу к порту addr
func NewApp(addr string) (*App, error) {
	// БД
	storageTask, err := task.NewSQLiteDB()
	if err != nil {
		logrus.Error(err)
		return nil, errors.Wrap(err, "Open SQLite DB")
	}
	// сервис
	serviceTask := task.NewService(storageTask)
	// хендлеры
	route := chi.NewRouter()
	handlerTask := NewHandler(serviceTask)
	// регистрация хендлеров в роутере chi
	handlerTask.Register(route)

	return &App{
		server: &http.Server{
			Addr:    strings.Join([]string{":", addr}, ""),
			Handler: route,
		},
		storage: storageTask,
		quitCh:  make(chan struct{}),
	}, nil
}

// Start запускает приложение на порту, переданном при инициализации в NewApp()
func (s *App) Start() error {
	var (
		errApp error // ошибка возникающия при работе сервера
	)
	// создание БД и миграции
	if err := s.createAndMigrateDb(); err != nil {
		return errors.Wrap(err, "checkAndMigrateDb method failed")
	}
	// горутина для запуска сервера
	go func() {
		logrus.WithFields(logrus.Fields{
			"ListenAddr": s.server.Addr,
		}).Info("server start")
		if err := s.server.ListenAndServe(); err != http.ErrServerClosed {
			logrus.Error(err)
			errApp = err
			close(s.quitCh)
		}
	}()
	// канал, ждущий завершения работы приложения
	<-s.quitCh
	return errApp
}

// Stop останавливает работу сервиса
func (s *App) Stop() {
	logrus.Debug("stop signal registered")
	// завершение работы сервера
	if err := s.server.Shutdown(context.Background()); err != nil {
		logrus.Errorf("stop server: %s", err.Error())
	}
	logrus.Debug("server stopped")
	// отключение от БД
	if err := s.storage.DissconecteDB(); err != nil {
		logrus.Errorf("close DB: %s", err.Error())
	}
	logrus.Debug("DB closed successfully")
	logrus.Info("App stopped")
	close(s.quitCh)
}

// createAndMigrateDb создает БД и применяет миграции к ней
func (s *App) createAndMigrateDb() error {
	pathToDB := os.Getenv("TODO_DBFILE")
	// открытие БД для дальнейших миграций
	db, err := goose.OpenDBWithDriver("sqlite", pathToDB)
	if err != nil {
		logrus.Errorf("DB open to migrate: %s", err.Error())
		return errors.Wrap(err, "DB open to migrate")
	}
	logrus.Debug("DB opened for migration successfully")

	defer func() {
		if err := db.Close(); err != nil {
			logrus.Errorf("DB close: %s", err.Error())
			return
		}
		logrus.Debug("DB for migration closed successfully")
	}()
	// проверка на уже существующие миграции БД
	// метод возвращает id миграции и nil в случае
	// уже существующих миграций
	_, err = goose.EnsureDBVersion(db)
	if err == nil {
		logrus.Debug("Migration of DB done already")
		return nil
	}
	// Выполняем миграции
	err = goose.Up(db, "./migrations")
	if err != nil {
		logrus.Errorf("DB migration: %s", err.Error())
		return err
	}
	logrus.Debug("Migration Up of DB successfully done")

	return nil
}
