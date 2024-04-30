package task

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
