package api

import (
	"net/http"
	"time"

	"github.com/VoC925/go_final_project/service/internal/domain/task"
	errorsApp "github.com/VoC925/go_final_project/service/internal/error_app"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

const (
	pathToHTMLFile = "./"
)

type handleRegister interface {
	Register(route *chi.Mux)
}

var _ handleRegister = &handleScheduler{}

type handleScheduler struct{}

func NewHandler() handleRegister {
	return &handleScheduler{}
}

func (h *handleScheduler) Register(route *chi.Mux) {
	route.Get("/", h.getHTMLPage)
	route.Get("/api/nextdate", h.nextDateSchedule) // "api/nextdate?now=20240126&date=20240126&repeat=y"
}

func (h *handleScheduler) getHTMLPage(w http.ResponseWriter, req *http.Request) {
	fs := http.FileServer(http.Dir(pathToHTMLFile))
	http.StripPrefix("/", fs).ServeHTTP(w, req)
}

func (h *handleScheduler) nextDateSchedule(w http.ResponseWriter, req *http.Request) {
	// парсинг параметров запроса
	var queryParams queryNextDateParams
	if err := queryParams.parsingFromQuery(req); err != nil {
		errorsApp.RequestError(w, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "parsingFromQuery() method"),
		))
		return
	}
	// вычисление следующей дачи задачи
	nextDateOfTask, err := task.NextDate(
		queryParams.Now,
		queryParams.Date,
		queryParams.Repeat,
	)
	if err != nil {
		errorsApp.RequestError(w, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "NextDate() method"),
		))
		return
	}
	// dataJson, err := json.Marshal(usersFromDB)
	// if err != nil {
	// 	requestError(w, apperror.New(http.StatusInternalServerError, err.Error()))
	// 	return
	// }
	errorsApp.RequestOk(w)
	w.Write([]byte(nextDateOfTask))
}

// структура для хранения параметров запроса
type queryNextDateParams struct {
	Now    time.Time
	Date   time.Time
	Repeat string
}

func (q *queryNextDateParams) parsingFromQuery(r *http.Request) error {
	var params queryNextDateParams
	// параметр now
	nowQuery := r.FormValue("now")
	if nowQuery == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "now")
	}
	nowQueryAsTime, err := time.Parse("20060102", nowQuery)
	if err != nil {
		return errors.Wrap(err, errorsApp.ErrInvalidData.Error())
	}
	params.Now = nowQueryAsTime
	// параметр date
	dateQuery := r.FormValue("date")
	if dateQuery == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "date")
	}
	dateQueryAsTime, err := time.Parse("20060102", dateQuery)
	if err != nil {
		return errors.Wrap(err, errorsApp.ErrInvalidData.Error())
	}
	params.Date = dateQueryAsTime
	// параметр repeat
	repeatQuery := r.FormValue("repeat")
	if repeatQuery == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "repeat")
	}
	params.Repeat = repeatQuery
	return nil
}
