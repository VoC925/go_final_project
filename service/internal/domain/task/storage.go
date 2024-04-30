package task

import (
	"context"
	"database/sql"
	"os"
)

// интерфейс БД
type Storage interface {
	Insert(context.Context, *CreateTaskDTO) (int, error)
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

func (s *storageTask) Insert(ctx context.Context, task *CreateTaskDTO) (int, error) {
	q := `INSERT INTO scheduler (date, title, comment, repeat)
VALUES (:dateVal, :titleVal, :commentVal, :repeatVal)`
	res, err := s.db.ExecContext(ctx, q,
		sql.Named("dateVal", task.Date),
		sql.Named("titleVal", task.Title),
		sql.Named("commentVal", task.Comment),
		sql.Named("repeatVal", task.Repeat))
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}
