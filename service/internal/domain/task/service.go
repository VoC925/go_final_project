package task

import (
	"context"
	"fmt"

	errorsApp "github.com/VoC925/go_final_project/service/internal/error_app"
	"github.com/pkg/errors"
)

// интерфейс сервиса для CRUD операций
type Service interface {
	// поиск по параметру
	FindTaskByParam(context.Context, string) (*Task, error)
	// Поиск задач
	FindTasks(ctx context.Context, offset int, limit int, search string) ([]*Task, error)
	// Добавление новой задачи
	InsertNewTask(context.Context, *CreateTaskDTO) (*Task, error)
	// // Удаление
	// Delete(context.Context, string) error
	// Изменение данных существующей задачи Task
	UpdateTask(context.Context, *Task) error
}

// структура, реализующая интерфейс Service
type serviceTask struct {
	// база данных
	db Storage
}

// Конструктор
func NewService(db Storage) Service {
	return &serviceTask{
		db: db,
	}
}

// InsertNewTask метод для добавления новой задачи
func (s *serviceTask) InsertNewTask(ctx context.Context, dto *CreateTaskDTO) (*Task, error) {
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

// FindTasks метод для поиска задач
func (s *serviceTask) FindTasks(ctx context.Context, offset int, limit int, search string) ([]*Task, error) {
	// запрос к БД для поиска задач
	tasks, err := s.db.Find(ctx, offset, limit, search)
	if err != nil {
		return nil, errors.Wrap(err, "find in DB")
	}
	if tasks == nil {
		return []*Task{}, nil
	}
	return tasks, nil
}

// FindTaskByParam метод для поиска задачи по параметру
func (s *serviceTask) FindTaskByParam(ctx context.Context, param string) (*Task, error) {
	// запрос к БД для поиска задачи по параметру
	task, err := s.db.FindByParamID(ctx, param)
	if errors.Is(err, errorsApp.ErrNoData) {
		return nil, fmt.Errorf("задача не найдена")
	}
	if err != nil {
		return nil, errors.Wrap(err, "find in DB")
	}
	return task, nil
}

func (s *serviceTask) UpdateTask(ctx context.Context, task *Task) error {
	// валидация данных структуры Task
	if err := task.Validate(); err != nil {
		return errors.Wrap(err, "validate request data")
	}
	// запрос к БД для обновления данных задачи
	err := s.db.Update(ctx, task)
	if errors.Is(err, errorsApp.ErrNoData) {
		return fmt.Errorf("задача не найдена")
	}
	if err != nil {
		return errors.Wrap(err, "update in DB")
	}
	return nil
}
