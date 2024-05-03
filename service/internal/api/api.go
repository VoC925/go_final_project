package api

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

	"github.com/VoC925/go_final_project/service/internal/domain/task"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

type App struct {
	server *http.Server
	quitCh chan struct{}
}

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
	handlerTask.Register(route)

	return &App{
		server: &http.Server{
			Addr:    strings.Join([]string{":", addr}, ""),
			Handler: route,
		},
		quitCh: make(chan struct{}),
	}, nil
}

func (s *App) Start() error {
	var (
		errApp error
	)

	if err := s.createAndMigrateDb(); err != nil {
		return errors.Wrap(err, "checkAndMigrateDb method failed")
	}

	go func() {
		logrus.WithFields(logrus.Fields{
			"ListenAddr": s.server.Addr,
		}).Info("server start")
		if err := s.server.ListenAndServe(); err != nil {
			errApp = err
			close(s.quitCh)
		}
	}()

	<-s.quitCh

	return errApp
}

// Stop останавливает работу сервиса
func (s *App) Stop() {
	close(s.quitCh)
}

func (s *App) createAndMigrateDb() error {
	pathToDB := os.Getenv("TODO_DBFILE")
	dbFile, err := sql.Open("sqlite", pathToDB)
	if err != nil {
		return errors.Wrap(err, "Open SQLite DB")
	}
	defer dbFile.Close()

	// создание таблицы tasks
	res, err := dbFile.Exec(`CREATE TABLE IF NOT EXISTS scheduler (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	date VARCHAR(8) NOT NULL DEFAULT "",
	title VARCHAR(128) NOT NULL DEFAULT "",
	comment TEXT NOT NULL DEFAULT "",
	repeat VARCHAR(128) NOT NULL DEFAULT ""
);
CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler (date);`)

	if err != nil {
		return errors.Wrap(err, "DB Query")
	}
	r, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if r != 0 {
		logrus.Info("Migration Up of DB successfully done")
	}
	return nil
}
