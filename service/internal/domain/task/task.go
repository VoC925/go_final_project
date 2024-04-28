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
	repeatVal, err := splitReapeatValue(repeat)
	if err != nil {
		return "", errors.Wrap(err, "splitReapeatValue() method")
	}

	switch repeatVal.typeRepeat {
	case "d":
		for {
			newDate := date.AddDate(0, 0, repeatVal.period)
			if newDate.After(now) {
				return date.Format("20060102"), nil
			}
		}
	case "y":
		for {
			newYear := date.AddDate(repeatVal.period, 0, 0)
			if newYear.After(now) {
				return date.Format("20060102"), nil
			}
		}
	default:
		return "", fmt.Errorf("error with param: repeat")
	}
}

func splitReapeatValue(repeat string) (*repeatTasks, error) {
	parts := strings.Split(repeat, " ")
	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return nil, fmt.Errorf("некорректно задан парамметр repeat")
		}
		repeat_days, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, err
		}
		if repeat_days > 400 {
			return nil, fmt.Errorf("число дней не может быть больше 400")
		}
		return &repeatTasks{
			typeRepeat: parts[0],
			period:     repeat_days,
		}, nil
	case "y":
		return &repeatTasks{
			typeRepeat: parts[0],
			period:     1,
		}, nil
	default:
		return nil, fmt.Errorf("неожиданный параметр repeat")
	}
}
