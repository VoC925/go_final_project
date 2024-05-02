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

// структура задачи `Data Transfer Object`
// для передачи при создании новой задачи
type CreateTaskDTO struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// UnmarshalJSONToStruct десериализует данные из JSON в структуру TaskDTO
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

// Структура задачи
type Task struct {
	ID      string `json:"id"`      // id задачи
	Date    string `json:"date"`    // дата выполнения задачи
	Title   string `json:"title"`   // название задачи
	Comment string `json:"comment"` // дополнительный текст задачи
	Repeat  string `json:"repeat"`  // периодичность выполнения задачи
}

// UnmarshalJSONToStruct десериализует данные из JSON в структуру Task
func (t *Task) UnmarshalJSONToStruct(data []byte) error {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return err
	}
	t.ID = task.ID
	t.Date = task.Date
	t.Title = task.Title
	t.Comment = task.Comment
	t.Repeat = task.Repeat
	return nil
}

// валидация данных структуры Task
func (t *Task) Validate() error {
	// валидация поля ID
	if t.ID == "" {
		return errors.Wrap(errorsApp.ErrEmptyField, "id")
	}
	if _, err := strconv.Atoi(t.ID); err != nil {
		return errors.Wrap(errorsApp.ErrInvalidData, "id")
	}
	d := &decorator{
		&CreateTaskDTO{
			Date:    t.Date,
			Title:   t.Title,
			Comment: t.Comment,
			Repeat:  t.Repeat,
		},
	}
	// валидация  остальных полей структуры Task
	if err := d.validateStruct(t); err != nil {
		return err
	}
	return nil
}

// структура обертка, для использования
// готового метода Validate() у структуры TaskDTO
type decorator struct {
	*CreateTaskDTO
}

// validateStruct валидирует поля date и title структуры Task
// используя метод структуры CreateTaskDTO через структуру обертку decorator
func (d *decorator) validateStruct(task *Task) error {
	// валидация title и date
	if err := d.Validate(); err != nil {
		return err
	}
	task.Date = d.Date
	// валидация repeat
	if d.Repeat != "" {
		if _, err := defineReapeatValue(d.Repeat); err != nil {
			return err
		}
	}
	return nil
}

// createTaskFromCreateTaskDTO создает структуру Task на основе TaskDTO
func createTaskFromCreateTaskDTO(dto *CreateTaskDTO) *Task {
	return &Task{
		ID:      "",
		Date:    dto.Date,
		Title:   dto.Title,
		Comment: dto.Comment,
		Repeat:  dto.Repeat,
	}
}
