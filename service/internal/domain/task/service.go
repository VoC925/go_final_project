package task

import (
	"context"

	"github.com/pkg/errors"
)

// интерфейс сервиса для CRUD операций
type Service interface {
	// // поиск по ID
	// FindByID(context.Context, string) (*User, error)
	// // поиск по email
	// GetByEmail(context.Context, string) (*User, error)
	// // Поиск всех пользователей
	// FindAll(context.Context, QueryParams) ([]*User, error)
	// Добавление новой задачи
	InsertNewTask(context.Context, *CreateTaskDTO) (*Task, error)
	// // Удаление
	// Delete(context.Context, string) error
	// // Update
	// Update(context.Context, string, []byte) error
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
