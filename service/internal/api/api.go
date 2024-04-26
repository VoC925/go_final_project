package api

import (
	"database/sql"
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

type App struct {
	server *http.Server
	quitCh chan struct{}
}

func NewApp(addr string, route http.Handler) *App {
	return &App{
		server: &http.Server{
			Addr:    strings.Join([]string{":", addr}, ""),
			Handler: route,
		},
		quitCh: make(chan struct{}),
	}
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
			s.Stop()
		}
	}()

	<-s.quitCh

	return errApp
}

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
	_, err = dbFile.Exec(`CREATE TABLE IF NOT EXISTS scheduler (
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

	logrus.Info("Migration Up of DB successfully done")

	return nil
}
