package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

const (
	pathToFile = "./"
)

type handleRegister interface {
	Register(route *chi.Mux)
}

var _ handleRegister = &handleUser{}

type handleUser struct {
	// service user.Service
}

func NewHandler() handleRegister {
	return &handleUser{}
}

func (h *handleUser) Register(route *chi.Mux) {
	route.Group(func(r chi.Router) {
		route.Get("/*", h.getHTMLPage)
	})
}

func (h *handleUser) getHTMLPage(w http.ResponseWriter, req *http.Request) {
	fs := http.FileServer(http.Dir(pathToFile))
	http.StripPrefix("/", fs).ServeHTTP(w, req)
}
