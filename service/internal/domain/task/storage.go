package task

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/VoC925/go_final_project/service/internal/httpResponse"
	"github.com/sirupsen/logrus"
)

// интерфейс БД
type Storage interface {
	// добавление в БД
	Insert(context.Context, *CreateTaskDTO) (string, error)
	// поиск в БД по параметрам
	Find(ctx context.Context, offset int, limit int, search string) ([]*Task, error)
	// поиск в БД по ID
	FindByParamID(context.Context, string) (*Task, error)
	// обновление в БД
	Update(context.Context, *Task) error
	// удаление в БД по ID
	Delete(context.Context, string) error
	// отключение соединения БД
	DissconecteDB() error
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

	logrus.WithFields(logrus.Fields{
		"path": pathToDB,
	}).Debug("sqlite DB connected")

	return &storageTask{
		db: dbFile,
	}, nil
}

// DissconecteDB разрывает текущее соединение к БД
func (s *storageTask) DissconecteDB() error {
	if err := s.db.Close(); err != nil {
		return err
	}
	return nil
}

// Insert метод добавления новой задачи в БД
func (s *storageTask) Insert(ctx context.Context, task *CreateTaskDTO) (string, error) {
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
		"task":     task.String(),
	}).Debug("sqlite request INSERT")

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
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
		"task id":  id,
	}).Debug("sqlite request INSERT success")
	return fmt.Sprint(id), nil
}

// Find метод для поиска задач в БД
func (s *storageTask) Find(ctx context.Context, offset int, limit int, search string) ([]*Task, error) {
	var (
		q     string    // SQL запрос
		rows  *sql.Rows // результат запроса
		tasks []*Task   // слайс задач
	)
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
		"offset":   offset,
		"limit":    limit,
		"search":   search,
	}).Debug("sqlite request SELECT")

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
	logrus.WithFields(logrus.Fields{
		"log_uuid":    ctx.Value("log_uuid").(string),
		"found tasks": len(tasks),
	}).Debug("sqlite request SELECT success")
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

// FindByParamID поиск задачи по ID
func (s *storageTask) FindByParamID(ctx context.Context, id string) (*Task, error) {
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
		"id":       id,
	}).Debug("sqlite request SELECT")
	q := `SELECT *
FROM scheduler
WHERE id = :paramVal;`
	row := s.db.QueryRowContext(ctx, q, sql.Named("paramVal", id))
	t := new(Task)
	err := row.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err == sql.ErrNoRows {
		return nil, httpResponse.ErrNoData
	}
	if err != nil {
		return nil, err
	}
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
	}).Debug("sqlite request SELECT success")
	return t, nil
}

// Update обновление задачи
func (s *storageTask) Update(ctx context.Context, task *Task) error {
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
		"task":     task.String(),
	}).Debug("sqlite request UPDATE")
	q := `UPDATE scheduler
SET date = :dateVal, title = :titleVal, comment = :commentVal, repeat = :repeatVal 
WHERE id = :idVal;`
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
		return httpResponse.ErrNoData
	}
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
	}).Debug("sqlite request UPDATE success")
	return nil
}

// Delete удаление задачи из БД
func (s *storageTask) Delete(ctx context.Context, id string) error {
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
		"id":       id,
	}).Debug("sqlite request DELETE")
	q := `DELETE FROM scheduler WHERE id=:idVal;`
	res, err := s.db.ExecContext(ctx, q, sql.Named("idVal", id))
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return httpResponse.ErrNoData
	}
	logrus.WithFields(logrus.Fields{
		"log_uuid": ctx.Value("log_uuid").(string),
	}).Debug("sqlite request DELETE success")
	return nil
}
