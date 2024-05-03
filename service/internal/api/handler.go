package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/VoC925/go_final_project/service/internal/domain/task"
	"github.com/VoC925/go_final_project/service/internal/httpResponse"
	"github.com/VoC925/go_final_project/service/pkg"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	webDir = "./web" // путь до фронтенда
)

type handleRegister interface {
	Register(route *chi.Mux)
}

var _ handleRegister = &handleScheduler{}

type handleScheduler struct {
	service task.Service
}

func NewHandler(s task.Service) handleRegister {
	logrus.Debug("handler Task creted")
	return &handleScheduler{
		service: s,
	}
}

func (h *handleScheduler) Register(route *chi.Mux) {
	// загрузка фронтенда
	route.Handle("/*", http.FileServer(http.Dir(webDir)))
	// r.Group(func(r chi.Router) {
	//     r.Use(AuthMiddleware)
	//     r.Post("/manage", CreateAsset)
	// })
	// аутентификация пользователя
	route.Post("/api/sign", h.authUser)
	route.Route("/api", func(r chi.Router) {
		// получение следующей даты задачи
		r.Get("/nextdate", h.nextDateSchedule)
		// добавление задачи
		r.Post("/task", h.handleAddTask)
		// получения списка задач
		r.Get("/tasks", h.handleGetTasks)
		// получение задачи по id
		r.Get("/task", h.handleGetTaskByID)
		// изменение существующей задачи
		r.Put("/task", h.handleUpdateTask)
		// заверешение существующей задачи
		r.Post("/task/done", h.handleTaskDone)
		// удаление существующей задачи
		r.Delete("/task", h.handleDeleteTask)
	})

}

// nextDateSchedule обработчик для получения следующей даты задачи
func (h *handleScheduler) nextDateSchedule(w http.ResponseWriter, req *http.Request) {
	var (
		cid       = uuid.New().String() // уникальный id для логов
		startTime = time.Now()          // время, относительно которого считается время выполнения запроса
	)
	// парсинг параметров запроса
	var queryParams queryNextDateParams
	if err := queryParams.parsingFromQuery(req); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(err, "parsingFromQuery() method"),
			),
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
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "NextDate() method"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid, req, []byte(nextDateOfTask), time.Since(startTime), nil))
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
		return errors.Wrap(httpResponse.ErrEmptyField, "now")
	}
	nowQueryAsTime, err := time.Parse("20060102", nowQuery)
	if err != nil {
		return errors.Wrap(err, httpResponse.ErrInvalidData.Error())
	}
	q.Now = nowQueryAsTime
	// параметр date
	dateQuery := r.FormValue("date")
	if dateQuery == "" {
		return errors.Wrap(httpResponse.ErrEmptyField, "date")
	}
	dateQueryAsTime, err := time.Parse("20060102", dateQuery)
	if err != nil {
		return errors.Wrap(err, httpResponse.ErrInvalidData.Error())
	}
	q.Date = dateQueryAsTime
	// параметр repeat
	repeatQuery := r.FormValue("repeat")
	if repeatQuery == "" {
		return errors.Wrap(httpResponse.ErrEmptyField, "repeat")
	}
	q.Repeat = repeatQuery
	return nil
}

// handleAddTask обработчик для добавления новой задачи
func (h *handleScheduler) handleAddTask(w http.ResponseWriter, req *http.Request) {
	var (
		buf       bytes.Buffer
		cid       = uuid.New() // уникальный id для логов
		ctx       = req.Context()
		startTime = time.Now() // время, относительно которого считается время выполнения запроса
	)
	// Чтение JSON из тела запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(err, "Read from request body"),
			),
		))
		return
	}
	defer req.Body.Close()
	// указатель на структуру новой задачи TaskDTO
	taskDTO := new(task.CreateTaskDTO)
	if err := taskDTO.UnmarshalJSONToStruct(buf.Bytes()); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "Unmarshal JSON"),
			),
		))
		return
	}
	// сервис
	taskInserted, err := h.service.InsertNewTask(ctx, taskDTO)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "service task"),
			),
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
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "marshal JSON"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid.String(), req, jsonData, time.Since(startTime), nil))
}

// handleGetTasks обработчик для получения задач
func (h *handleScheduler) handleGetTasks(w http.ResponseWriter, req *http.Request) {
	var (
		ctx         = req.Context()
		queryParams queryGetTaskParams
		cid         = uuid.New().String() // уникальный id для логов
		startTime   = time.Now()          // время, относительно которого считается время выполнения запроса
	)
	// парсинг параметров запроса
	if err := parseQueryParamsGetTasks(&queryParams, req); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(err, "parse query parametrs"),
			),
		))
		return
	}
	// сервис
	tasks, err := h.service.FindTasks(ctx, queryParams.Offest, queryParams.Limit, queryParams.Search)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "service task"),
			),
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
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "marshal JSON"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid, req, jsonData, time.Since(startTime), nil))
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
			return errors.Wrap(httpResponse.ErrInvalidQueryParams, "limit < 0")
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
			return errors.Wrap(httpResponse.ErrInvalidQueryParams, "offset < 0")
		}
		dest.Offest = offs
	}
	// параметр search
	dest.Search = r.FormValue("search")
	return nil
}

// handleGetTaskByID обработчик для получения задачи по ID
func (h *handleScheduler) handleGetTaskByID(w http.ResponseWriter, req *http.Request) {
	var (
		ctx       = req.Context()
		cid       = uuid.New().String() // уникальный id для логов
		startTime = time.Now()          // время, относительно которого считается время выполнения запроса
	)
	// парсинг ID парметра
	idQuery := req.FormValue("id")
	if idQuery == "" {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(httpResponse.ErrInvalidQueryParams, "ID"),
			),
		))
		return
	}
	// сервис
	task, err := h.service.FindTaskByParam(ctx, idQuery)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "service task"),
			),
		))
		return
	}
	// получение JSON данных ответа, содержащего слайс задач
	jsonData, err := json.Marshal(task)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "marshal JSON"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid, req, jsonData, time.Since(startTime), nil))
}

// handleUpdateTask обработчик для обновления задачи
func (h *handleScheduler) handleUpdateTask(w http.ResponseWriter, req *http.Request) {
	var (
		buf       bytes.Buffer
		ctx       = req.Context()
		cid       = uuid.New().String() // уникальный id для логов
		startTime = time.Now()          // время, относительно которого считается время выполнения запроса
	)
	// Чтение JSON из тела запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(err, "Read from request body"),
			),
		))
		return
	}
	defer req.Body.Close()
	// указатель на структуру новой задачи TaskDTO
	task := new(task.Task)
	if err := task.UnmarshalJSONToStruct(buf.Bytes()); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "Unmarshal JSON"),
			),
		))
		return
	}
	// сервис
	if err := h.service.UpdateTask(ctx, task); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "service task"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid, req, []byte(`{}`), time.Since(startTime), nil))
}

// handleTaskDone обработчик для завершения задачи
func (h *handleScheduler) handleTaskDone(w http.ResponseWriter, req *http.Request) {
	var (
		ctx       = req.Context()
		cid       = uuid.New().String() // уникальный id для логов
		startTime = time.Now()          // время, относительно которого считается время выполнения запроса
	)
	// парсинг ID параметра
	idQuery := req.FormValue("id")
	if idQuery == "" {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(httpResponse.ErrInvalidQueryParams, "ID"),
			),
		))
		return
	}
	// сервис
	if err := h.service.TaskDone(ctx, idQuery); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "service task"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid, req, []byte(`{}`), time.Since(startTime), nil))
}

// handleDeleteTask обработчик для удаления задачи
func (h *handleScheduler) handleDeleteTask(w http.ResponseWriter, req *http.Request) {
	var (
		ctx       = req.Context()
		cid       = uuid.New().String() // уникальный id для логов
		startTime = time.Now()          // время, относительно которого считается время выполнения запроса
	)
	// парсинг ID параметра
	idQuery := req.FormValue("id")
	if idQuery == "" {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(httpResponse.ErrInvalidQueryParams, "ID"),
			),
		))
		return
	}
	// сервис
	if err := h.service.DeleteTask(ctx, idQuery); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid, req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "service task"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid, req, []byte(`{}`), time.Since(startTime), nil))
}

// authUser аутентификация полльзователя по паролю
func (h *handleScheduler) authUser(w http.ResponseWriter, req *http.Request) {
	var (
		buf bytes.Buffer
		cid = uuid.New() // уникальный id для логов
		// ctx       = req.Context()
		startTime = time.Now() // время, относительно которого считается время выполнения запроса
	)
	// Чтение JSON из тела запроса
	_, err := buf.ReadFrom(req.Body)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusBadRequest,
				errors.Wrap(err, "Read from request body"),
			),
		))
		return
	}
	defer req.Body.Close()
	// анонимная структура для хранения пароля из запроса
	password := struct {
		Val string `json:"password"`
	}{}
	// десериализация пароля
	if err := json.Unmarshal(buf.Bytes(), &password); err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "Unmarshal JSON"),
			),
		))
		return
	}
	// проверка пароля
	if password.Val != os.Getenv("TODO_PASSWORD") {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusUnauthorized,
				httpResponse.ErrUnAuth,
			),
		))
		return
	}
	// создание токена
	tokenStr, err := pkg.CreateToken()
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "Authorization"),
			),
		))
		return
	}
	// анонимная структура, содержащая token
	tokenResponse := struct {
		Val string `json:"token"`
	}{
		Val: tokenStr,
	}
	// получение JSON данных ответа, содержащего токен
	jsonData, err := json.Marshal(tokenResponse)
	if err != nil {
		httpResponse.Error(w, httpResponse.NewLogInfo(cid.String(), req, nil, time.Since(startTime),
			httpResponse.NewError(
				http.StatusInternalServerError,
				errors.Wrap(err, "marshal JSON"),
			),
		))
		return
	}

	httpResponse.Success(w, httpResponse.NewLogInfo(cid.String(), req, jsonData, time.Since(startTime), nil))
}
