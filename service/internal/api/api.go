package api

import (
	"net/http"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	pathToDB = "./scheduler.db"
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

	if err := s.checkAndMigrateDb(); err != nil {
		return errors.Wrap(err, "checkAndMigreteDb method failed")
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

func (s *App) createDBFile(path string) (*os.File, error) {
	dbFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, errors.Wrap(err, "os.OpenFile() failed")
	}
	return dbFile, nil
}

func (s *App) checkAndMigrateDb() error {
	// проверка существования файла ДБ
	if _, err := os.Stat(pathToDB); os.IsNotExist(err) {
		dbFile, err := s.createDBFile(pathToDB)
		if err != nil {
			return errors.Wrap(err, "createDBFile")
		}
		defer dbFile.Close()
	}
	return nil
}
