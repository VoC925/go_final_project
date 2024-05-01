package task

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	errorsApp "github.com/VoC925/go_final_project/service/internal/error_app"
	"github.com/pkg/errors"
)

// структура для разделения параметра repeat
// например, "d 7" -> typeRepeat = "d"; period = 7
// "y" -> typeRepeat = "y"; period = 1
type repeatTasks struct {
	typeRepeat string
	period     int
}

func NextDate(now time.Time, date time.Time, repeat string) (string, error) {
	// валидация repeat
	repeatVal, err := defineReapeatValue(repeat)
	if err != nil {
		return "", errors.Wrap(err, "splitReapeatValue() method")
	}

	var condition bool

	switch repeatVal.typeRepeat {
	case "d":
		for !condition {
			date = date.AddDate(0, 0, repeatVal.period)
			if date.After(now) {
				condition = true
			}
		}
	case "y":
		for !condition {
			date = date.AddDate(repeatVal.period, 0, 0)
			if date.After(now) {
				condition = true
			}
		}
	default:
		return "", fmt.Errorf("error with param: repeat")
	}
	return date.Format("20060102"), nil
}

func defineReapeatValue(repeat string) (*repeatTasks, error) {
	switch repeat[0] {
	case 'd':
		parts := strings.Split(repeat, " ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("некорректно задан параметр repeat в формате <d _>")
		}
		repeat_days, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if repeat_days <= 0 {
			return nil, fmt.Errorf("число дней не может быть 0 или отрицательным")
		}
		if repeat_days > 400 {
			return nil, fmt.Errorf("число дней не может быть больше 400")
		}
		return &repeatTasks{
			typeRepeat: parts[0],
			period:     repeat_days,
		}, nil
	case 'y':
		if len(repeat) != 1 {
			return nil, fmt.Errorf("некорректно задан параметр repeat в формате <y>")
		}
		return &repeatTasks{
			typeRepeat: "y",
			period:     1,
		}, nil
	default:
		return nil, fmt.Errorf("неподдерживаемый параметр repeat")
	}
}

// Структура задачи
type Task struct {
	ID      string `json:"id"`      // id задачи
	Date    string `json:"date"`    // дата выполнения задачи
	Title   string `json:"title"`   // название задачи
	Comment string `json:"comment"` // дополнительный текст задачи
	Repeat  string `json:"repeat"`  // периодичность выполнения задачи
}

// CreateTaskFromCreateTaskDTO создает структура на основе DTO
func createTaskFromCreateTaskDTO(dto *CreateTaskDTO) *Task {
	return &Task{
		ID:      "",
		Date:    dto.Date,
		Title:   dto.Title,
		Comment: dto.Comment,
		Repeat:  dto.Repeat,
	}
}

// структура задачи `Data Transfer Object`
type CreateTaskDTO struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// UnmarshalJSON десериализует данные из в структуру TaskDTO
func (q *CreateTaskDTO) UnmarshalJSONToStruct(data []byte) error {
	var taskDTO CreateTaskDTO
	if err := json.Unmarshal(data, &taskDTO); err != nil {
		return err
	}
	q.Date = taskDTO.Date
	q.Title = taskDTO.Title
	q.Comment = taskDTO.Comment
	q.Repeat = taskDTO.Repeat
	return nil
}

// Valiadate проверяет данные структуры
func (q *CreateTaskDTO) Validate() error {
	// поле Title
	if q.Title == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "title")
	}
	timeNow := time.Now()
	// поле Date
	if q.Date == "" {
		q.Date = timeNow.Format("20060102")
	} else {
		date, err := time.Parse("20060102", q.Date)
		if err != nil {
			return errors.Wrap(err, errorsApp.ErrInvalidData.Error())
		}
		if q.Date == timeNow.Format("20060102") {
			return nil
		}
		if date.Before(timeNow) {
			switch q.Repeat {
			case "":
				q.Date = timeNow.Format("20060102")
			default:
				dateNext, err := NextDate(timeNow, date, q.Repeat)
				if err != nil {
					return err
				}
				q.Date = dateNext
			}
		}
	}
	return nil
}
