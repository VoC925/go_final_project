package task

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	errorsApp "github.com/VoC925/go_final_project/service/internal/error_app"
)

// интерфейс БД
type Storage interface {
	Insert(context.Context, *CreateTaskDTO) (string, error)
	Find(ctx context.Context, offset int, limit int, search string) ([]*Task, error)
	FindByParamID(context.Context, string) (*Task, error)
	Update(context.Context, *Task) error
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
VALUES (:dateVal, :titleVal, :commentVal, :repeatVal);`
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
func (s *storageTask) Find(ctx context.Context, offset int, limit int, search string) ([]*Task, error) {
	var (
		q     string    // SQL запрос
		rows  *sql.Rows // результат запроса
		tasks []*Task   // слайс задач
	)

	if search == "" {
		// случай отсутствия параметра search
		q = `SELECT *
FROM scheduler
ORDER BY date
LIMIT :limitVal OFFSET :offsetVal;`
		r, err := s.db.QueryContext(ctx, q,
			sql.Named("limitVal", limit),
			sql.Named("offsetVal", offset))
		if err != nil {
			return nil, err
		}
		rows = r
	} else {
		// случай наличия параметра search
		time, isTime := searchIsTime(search)
		switch isTime {
		case true:
			// параметр search - дата
			search = time
			q = `SELECT *
FROM scheduler
WHERE date = :searchVal
LIMIT :limitVal;`
			r, err := s.db.QueryContext(ctx, q,
				sql.Named("searchVal", search),
				sql.Named("limitVal", limit))
			if err != nil {
				return nil, err
			}
			rows = r
		case false:
			// параметр search - поиск по title/comment
			q = `SELECT *
FROM scheduler
WHERE title LIKE '%' || :searchVal || '%' OR comment LIKE '%' || :searchVal || '%'
ORDER BY date
LIMIT :limitVal;`
			r, err := s.db.QueryContext(ctx, q,
				sql.Named("searchVal", search),
				sql.Named("limitVal", limit))
			if err != nil {
				return nil, err
			}
			rows = r
		}
	}

	defer rows.Close()

	for rows.Next() {
		t := Task{}
		err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
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

// searchIsTime определяет является ли запрос поиском даты
func searchIsTime(s string) (string, bool) {
	_, err := time.Parse("02.01.2006", s)
	if err == nil {
		var timePattern strings.Builder
		parts := strings.Split(s, ".")
		timePattern.WriteString(parts[2])
		timePattern.WriteString(parts[1])
		timePattern.WriteString(parts[0])
		return timePattern.String(), true
	}
	return "", false
}

func (s *storageTask) FindByParamID(ctx context.Context, id string) (*Task, error) {
	q := `SELECT *
FROM scheduler
WHERE id = :paramVal;`
	row := s.db.QueryRowContext(ctx, q, sql.Named("paramVal", id))
	t := new(Task)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err == sql.ErrNoRows {
		return nil, errorsApp.ErrNoData
	}
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (s *storageTask) Update(ctx context.Context, task *Task) error {
	q := `UPDATE scheduler
SET date = :dateVal, title = :titleVal, comment = :commentVal, repeat = :repeatVal 
WHERE id = :idVal`
	res, err := s.db.ExecContext(ctx, q,
		sql.Named("dateVal", task.Date),
		sql.Named("titleVal", task.Title),
		sql.Named("commentVal", task.Comment),
		sql.Named("repeatVal", task.Repeat),
		sql.Named("idVal", task.ID))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errorsApp.ErrNoData
	}
	return nil
}
