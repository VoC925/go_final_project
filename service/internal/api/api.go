package api

import (
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
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
	var errApp error
	go func() {
		logrus.WithFields(logrus.Fields{
			"ListenAddr": s.server.Addr,
		}).Info("server start")
		// http.Handle("/", http.FileServer(http.Dir(pathToHTMLFile)))
		if err := s.server.ListenAndServe(); err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("server error")
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
