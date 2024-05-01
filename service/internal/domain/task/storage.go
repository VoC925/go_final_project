package task

import (
	"context"
	"database/sql"
	"fmt"
	"os"
)

// интерфейс БД
type Storage interface {
	Insert(context.Context, *CreateTaskDTO) (string, error)
	Find(ctx context.Context, offset int, limit int) ([]*Task, error)
}

type storageTask struct {
	db *sql.DB
}

// Конструктор
func NewSQLiteDB() (Storage, error) {
	pathToDB := os.Getenv("TODO_DBFILE")
	dbFile, err := sql.Open("sqlite", pathToDB)
	if err != nil {
		return nil, err
	}
	return &storageTask{
		db: dbFile,
	}, nil
}

// Insert метод добавления новой задачи в БД
func (s *storageTask) Insert(ctx context.Context, task *CreateTaskDTO) (string, error) {
	q := `INSERT INTO scheduler (date, title, comment, repeat)
VALUES (:dateVal, :titleVal, :commentVal, :repeatVal)`
	res, err := s.db.ExecContext(ctx, q,
		sql.Named("dateVal", task.Date),
		sql.Named("titleVal", task.Title),
		sql.Named("commentVal", task.Comment),
		sql.Named("repeatVal", task.Repeat))
	if err != nil {
		return "", err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return "", err
	}
	return fmt.Sprint(id), nil
}

// Find метод для поиска задач в БД
func (s *storageTask) Find(ctx context.Context, offset int, limit int) ([]*Task, error) {
	q := `SELECT *
FROM scheduler
ORDER BY date
LIMIT :limitVal OFFSET :offsetVal`
	rows, err := s.db.QueryContext(ctx, q,
		sql.Named("limitVal", limit),
		sql.Named("offsetVal", offset))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task

	for rows.Next() {
		t := Task{}
		err = rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return tasks, nil
}
