package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/VoC925/go_final_project/service/internal/domain/task"
	errorsApp "github.com/VoC925/go_final_project/service/internal/error_app"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

const (
	pathToHTMLFile = "../web"
)

type handleRegister interface {
	Register(route *chi.Mux)
}

var _ handleRegister = &handleScheduler{}

type handleScheduler struct {
	service task.Service
}

func NewHandler(s task.Service) handleRegister {
	return &handleScheduler{
		service: s,
	}
}

func (h *handleScheduler) Register(route *chi.Mux) {
	route.Get("/*", h.getHTMLPage)
	route.Get("/api/nextdate", h.nextDateSchedule)
	route.Post("/api/task", h.handleAddTask)
}

// getHTMLPage обработчик для загрузки фронтенда
func (h *handleScheduler) getHTMLPage(w http.ResponseWriter, req *http.Request) {
	path := filepath.Join(filepath.Dir(os.Args[0]), pathToHTMLFile)
	fs := http.FileServer(http.Dir(path))
	http.StripPrefix("/", fs).ServeHTTP(w, req)
}

// nextDateSchedule обработчик для получения следующей даты задачи
func (h *handleScheduler) nextDateSchedule(w http.ResponseWriter, req *http.Request) {
	// парсинг параметров запроса
	var queryParams queryNextDateParams
	if err := queryParams.parsingFromQuery(req); err != nil {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusBadRequest,
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
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "NextDate() method"),
		))
		return
	}

	errorsApp.RequestOk(
		w,
		http.MethodGet,
		strings.NewReader(nextDateOfTask),
	)
}

// структура для хранения параметров запроса по обработчику nextDateSchedule
type queryNextDateParams struct {
	Now    time.Time
	Date   time.Time
	Repeat string
}

// parsingFromQuery метод для получения парметров из запроса
func (q *queryNextDateParams) parsingFromQuery(r *http.Request) error {
	// параметр now
	nowQuery := r.FormValue("now")
	if nowQuery == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "now")
	}
	nowQueryAsTime, err := time.Parse("20060102", nowQuery)
	if err != nil {
		return errors.Wrap(err, errorsApp.ErrInvalidData.Error())
	}
	q.Now = nowQueryAsTime
	// параметр date
	dateQuery := r.FormValue("date")
	if dateQuery == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "date")
	}
	dateQueryAsTime, err := time.Parse("20060102", dateQuery)
	if err != nil {
		return errors.Wrap(err, errorsApp.ErrInvalidData.Error())
	}
	q.Date = dateQueryAsTime
	// параметр repeat
	repeatQuery := r.FormValue("repeat")
	if repeatQuery == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "repeat")
	}
	q.Repeat = repeatQuery
	return nil
}

// handleAddTask обработчик для добавления новой задачи
func (h *handleScheduler) handleAddTask(w http.ResponseWriter, req *http.Request) {
	var (
		buf bytes.Buffer
		ctx = req.Context()
	)
	// Чтение JSON из тела запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		errorsApp.RequestError(w, http.MethodPost, errorsApp.NewError(
			http.StatusBadRequest,
			errors.Wrap(err, "Read from request body"),
		))
		return
	}
	defer req.Body.Close()
	// указатель на структуру новой задачи TaskDTO
	taskDTO := new(task.CreateTaskDTO)
	if err := taskDTO.UnmarshalJSONToStruct(buf.Bytes()); err != nil {
		errorsApp.RequestError(w, http.MethodPost, errorsApp.NewError(
			http.StatusBadRequest,
			errors.Wrap(err, "Unmarshal JSON"),
		))
		return
	}
	// сервис
	taskInserted, err := h.service.InsertNewTask(ctx, taskDTO)
	if err != nil {
		errorsApp.RequestError(w, http.MethodPost, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "service task"),
		))
		return
	}
	// анонимная структура, содержащая id созданной задачи
	idResponse := struct {
		Id int `json:"id"`
	}{
		Id: taskInserted.ID,
	}
	// получение JSON данных ответа, содержащего id созданной задачи
	jsonData, err := json.Marshal(idResponse)
	if err != nil {
		errorsApp.RequestError(w, http.MethodPost, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "marshal JSON"),
		))
		return
	}

	errorsApp.RequestOk(
		w,
		http.MethodPost,
		strings.NewReader(string(jsonData)),
	)
}
