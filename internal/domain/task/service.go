package task

import (
	"context"
	"fmt"
	"time"

	"github.com/VoC925/go_final_project/internal/httpResponse"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// интерфейс сервиса для CRUD операций
type Service interface {
	// поиск по параметру
	GetTask(context.Context, string) (*Task, error)
	// Поиск задач
	GetTasks(ctx context.Context, offset int, limit int, search string) ([]*Task, error)
	// Добавление новой задачи
	AddTask(context.Context, *CreateTaskDTO) (*Task, error)
	// Завершение задачи
	TaskDone(context.Context, string) error
	// Изменение данных существующей задачи Task
	UpdateTask(context.Context, *Task) error
	// Удаление существующей задачи
	DeleteTask(context.Context, string) error
}

// структура, реализующая интерфейс Service
type serviceTask struct {
	// база данных
	db Storage
}

// Конструктор
func NewService(db Storage) Service {
	logrus.Debug("service Task created")
	return &serviceTask{
		db: db,
	}
}

// AddTask метод для добавления новой задачи
func (s *serviceTask) AddTask(ctx context.Context, dto *CreateTaskDTO) (*Task, error) {
	// валидация данных task
	if err := dto.Validate(); err != nil {
		return nil, errors.Wrap(err, "validate request data")
	}
	// запрос к БД, для добавления новой задачи
	id, err := s.db.Insert(ctx, dto)
	if err != nil {
		return nil, errors.Wrap(err, "insert to DB")
	}
	// создание структуры новой задачи
	newTask := createTaskFromCreateTaskDTO(dto)
	newTask.ID = id
	return newTask, nil
}

// GetTasks метод для поиска задач
func (s *serviceTask) GetTasks(ctx context.Context, offset int, limit int, search string) ([]*Task, error) {
	// запрос к БД для поиска задач
	tasks, err := s.db.Find(ctx, offset, limit, search)
	if err != nil {
		return nil, errors.Wrap(err, "find in DB")
	}
	if tasks == nil {
		logrus.Debug("no data found in DB")
		return []*Task{}, nil
	}
	return tasks, nil
}

// GetTask метод для поиска задачи по параметру
func (s *serviceTask) GetTask(ctx context.Context, param string) (*Task, error) {
	// запрос к БД для поиска задачи по параметру
	task, err := s.db.FindByParamID(ctx, param)
	if errors.Is(err, httpResponse.ErrNoData) {
		return nil, fmt.Errorf("задача не найдена")
	}
	if err != nil {
		return nil, errors.Wrap(err, "find in DB")
	}
	return task, nil
}

// UpdateTask метод для обновления текущей задачи
func (s *serviceTask) UpdateTask(ctx context.Context, task *Task) error {
	// валидация данных структуры Task
	if err := task.Validate(); err != nil {
		return errors.Wrap(err, "validate request data")
	}
	// запрос к БД для обновления данных задачи
	err := s.db.Update(ctx, task)
	if errors.Is(err, httpResponse.ErrNoData) {
		return fmt.Errorf("задача не найдена")
	}
	if err != nil {
		return errors.Wrap(err, "update in DB")
	}
	return nil
}

// TaskDone метод для завершения текущей задачи
func (s *serviceTask) TaskDone(ctx context.Context, id string) error {
	// получение задачи по ID из БД
	task, err := s.GetTask(ctx, id)
	if err != nil {
		return err
	}
	// удаление задачи при пустом поле repeat
	if task.Repeat == "" {
		if err := s.DeleteTask(ctx, task.ID); err != nil {
			return err
		}
		return nil
	}
	date, err := time.Parse("20060102", task.Date)
	if err != nil {
		return errors.Wrap(err, httpResponse.ErrInvalidData.Error())
	}
	// обновление даты завершения
	nextDate, err := NextDate(time.Now(), date, task.Repeat)
	if err != nil {
		return errors.Wrap(err, "NextDate()")
	}
	task.Date = nextDate
	// обновление задачи в БД
	if err := s.UpdateTask(ctx, task); err != nil {
		return err
	}
	return nil
}

// DeleteTask метод для удаления текущей задачи
func (s *serviceTask) DeleteTask(ctx context.Context, id string) error {
	// запрос к БД на удаление
	err := s.db.Delete(ctx, id)
	if errors.Is(err, httpResponse.ErrNoData) {
		return fmt.Errorf("задача не найдена")
	}
	if err != nil {
		return errors.Wrap(err, "delete in DB")
	}
	return nil
}
