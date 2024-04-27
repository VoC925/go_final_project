package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
)

const (
	pathToHTMLFile = "./"
)

type handleRegister interface {
	Register(route *chi.Mux)
}

var _ handleRegister = &handleUser{}

type handleUser struct {
	// service user.Service
}

func NewHandler() handleRegister {
	return &handleUser{}
}

func (h *handleUser) Register(route *chi.Mux) {
	route.Group(func(r chi.Router) {
		route.Get("/*", h.getHTMLPage)
	})
}

func (h *handleUser) getHTMLPage(w http.ResponseWriter, req *http.Request) {
	fs := http.FileServer(http.Dir(pathToHTMLFile))
	http.StripPrefix("/", fs).ServeHTTP(w, req)
}

var (
	ErrEmptyField  = errors.New("задано пустое поле")
	ErrInvalidData = errors.New("некорректное поле")
)

type repeatTasks struct {
	typeRepeat string
	period     int
}

func nextDate(now time.Time, date string, repeat string) (string, error) {
	// валидация repeat
	if repeat == "" {
		return "", errors.Wrap(ErrEmptyField, "repeat")
	}
	repeatVal, ok := validateReapeatValue(repeat)
	if !ok {
		return "", errors.Wrap(ErrInvalidData, "repeat")
	}
	// валидация date
	timeDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", errors.Wrap(err, ErrInvalidData.Error())
	}

	switch repeatVal.typeRepeat {
	case "d":
		for {
			timeDate.AddDate(0, 0, repeatVal.period)
			if now.After(timeDate) {
				return timeDate.Format("20060102"), nil
			}
		}
	case "y":
		for {
			timeDate.AddDate(repeatVal.period, 0, 0)
			if now.After(timeDate) {
				return timeDate.Format("20060102"), nil
			}
		}
	}
	return "", errors.Wrap(err, ErrInvalidData.Error())
}

func validateReapeatValue(repeat string) (*repeatTasks, bool) {
	parts := strings.Split(repeat, " ")
	switch parts[0] {
	case "d":
		if len(parts) != 2 {
			return nil, false
		}
		repeat_days, err := strconv.Atoi(parts[1])
		if err != nil || repeat_days > 400 {
			return nil, false
		}
		return &repeatTasks{
			typeRepeat: parts[0],
			period:     repeat_days,
		}, true
	case "y":
		return &repeatTasks{
			typeRepeat: parts[0],
			period:     1,
		}, true
	default:
		return nil, false
	}
}
