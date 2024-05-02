package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/VoC925/go_final_project/service/internal/domain/task"
	errorsApp "github.com/VoC925/go_final_project/service/internal/error_app"
	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

const (
	webDir = "./web"
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
	// загрузка фронтенда
	route.Handle("/*", http.FileServer(http.Dir(webDir)))
	// получение следующей даты задачи
	route.Get("/api/nextdate", h.nextDateSchedule)
	// добавление задачи
	route.Post("/api/task", h.handleAddTask)
	// получения списка задач
	route.Get("/api/tasks", h.handleGetTasks)
	// получение задачи по id
	route.Get("/api/task", h.handleGetTaskByID)
	// изменение существующей задачи
	route.Put("/api/task", h.handleUpdateTask)
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

// parsingFromQuery метод для получения параметров из запроса
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
		Id string `json:"id"`
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

func (h *handleScheduler) handleGetTasks(w http.ResponseWriter, req *http.Request) {
	var (
		ctx         = req.Context()
		queryParams queryGetTaskParams
	)
	// парсинг параметров запроса
	if err := parseQueryParamsGetTasks(&queryParams, req); err != nil {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusBadRequest,
			errors.Wrap(err, "parse query parametrs"),
		))
		return
	}
	// сервис
	tasks, err := h.service.FindTasks(ctx, queryParams.Offest, queryParams.Limit, queryParams.Search)
	if err != nil {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "service task"),
		))
		return
	}
	// анонимная структура, содержащая слайс Task
	tasksResponse := struct {
		Tasks []*task.Task `json:"tasks"`
	}{
		Tasks: tasks,
	}
	// получение JSON данных ответа, содержащего слайс задач
	jsonData, err := json.Marshal(tasksResponse)
	if err != nil {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "marshal JSON"),
		))
		return
	}

	errorsApp.RequestOk(
		w,
		http.MethodGet,
		strings.NewReader(string(jsonData)),
	)
}

// структура для хранения параметров запроса по обработчику handleGetTasks
type queryGetTaskParams struct {
	Limit  int    // количество записей на странице
	Offest int    // смещение записей
	Search string // поиск в строке
}

// parseQueryParamsGetTasks парсит параметры запроса из url.Values в структуру queryGetTaskParams
func parseQueryParamsGetTasks(dest *queryGetTaskParams, r *http.Request) error {
	// параметр limit
	limitQuery := r.FormValue("limit")
	if limitQuery == "" {
		dest.Limit = 10
	} else {
		lim, err := strconv.Atoi(limitQuery)
		if err != nil {
			return err
		}
		if lim < 0 {
			return errors.Wrap(errorsApp.ErrInvalidQueryParams, "limit < 0")
		}
		dest.Limit = lim
	}
	// параметр offset
	offsetQuery := r.FormValue("offset")
	if offsetQuery == "" {
		dest.Offest = 0
	} else {
		offs, err := strconv.Atoi(offsetQuery)
		if err != nil {
			return err
		}
		if offs < 0 {
			return errors.Wrap(errorsApp.ErrInvalidQueryParams, "offset < 0")
		}
		dest.Offest = offs
	}
	// параметр search
	dest.Search = r.FormValue("search")
	return nil
}

func (h *handleScheduler) handleGetTaskByID(w http.ResponseWriter, req *http.Request) {
	var (
		ctx = req.Context()
	)
	// парсинг ID парметра
	idQuery := req.FormValue("id")
	if idQuery == "" {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusBadRequest,
			errors.Wrap(errorsApp.ErrInvalidQueryParams, "ID"),
		))
		return
	}
	// сервис
	task, err := h.service.FindTaskByParam(ctx, idQuery)
	if err != nil {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "service task"),
		))
		return
	}
	// получение JSON данных ответа, содержащего слайс задач
	jsonData, err := json.Marshal(task)
	if err != nil {
		errorsApp.RequestError(w, http.MethodGet, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "marshal JSON"),
		))
		return
	}

	errorsApp.RequestOk(
		w,
		http.MethodGet,
		strings.NewReader(string(jsonData)),
	)
}

func (h *handleScheduler) handleUpdateTask(w http.ResponseWriter, req *http.Request) {
	var (
		buf bytes.Buffer
		ctx = req.Context()
	)
	// Чтение JSON из тела запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		errorsApp.RequestError(w, http.MethodPut, errorsApp.NewError(
			http.StatusBadRequest,
			errors.Wrap(err, "Read from request body"),
		))
		return
	}
	defer req.Body.Close()
	// указатель на структуру новой задачи TaskDTO
	task := new(task.Task)
	if err := task.UnmarshalJSONToStruct(buf.Bytes()); err != nil {
		errorsApp.RequestError(w, http.MethodPut, errorsApp.NewError(
			http.StatusBadRequest,
			errors.Wrap(err, "Unmarshal JSON"),
		))
		return
	}
	// сервис
	if err := h.service.UpdateTask(ctx, task); err != nil {
		errorsApp.RequestError(w, http.MethodPut, errorsApp.NewError(
			http.StatusInternalServerError,
			errors.Wrap(err, "service task"),
		))
		return
	}

	errorsApp.RequestOk(
		w,
		http.MethodPut,
		strings.NewReader(string(`{}`)),
	)
}
