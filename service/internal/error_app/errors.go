package errorsApp

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// общие ошибки
	ErrEmptyField  = errors.New("задано пустое поле")
	ErrInvalidData = errors.New("некорректное поле")
)

// интерфейс ошибок для HTTP ответов сервера
type HTTPError interface {
	// реализация интерфейса Stringer
	Error() string
	// маршалинг структуры ошибки в Json объект
	Marshal() ([]byte, error)
	// возвращение кода ошибки
	Status() int
	// добавить prefix
	WrapErr(error) error
}

// Структура, методы которой реализуюТ интерфейс HTTPError
type appError struct {
	Code int   `json:"code"`
	Err  error `json:"error"`
}

func (e appError) Error() string {
	return e.Err.Error()
}

func (e appError) Status() int {
	return e.Code
}

func (e appError) Marshal() ([]byte, error) {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (e appError) WrapErr(err error) error {
	return errors.Wrap(e, err.Error())
}

// конструктор структуры appError
func NewError(code int, err error) appError {
	return appError{Code: code, Err: err}
}

// Готовые ошибки
// func ErrInvalidID() appError {
// 	return appError{Code: http.StatusBadRequest, MsgError: "invalid ID"}
// }

// func ErrUnAuthorised() appError {
// 	return appError{Code: http.StatusUnauthorized, MsgError: "Unautorised"}
// }

// функция для формирования http заголовков для ответа клиенту
// в случае запроса, приведшего к ошибке
func RequestError(w http.ResponseWriter, errApp HTTPError) {
	w.Header().Set("Content-Type", "application/json")
	// добавление ошибки
	errApp.WrapErr(fmt.Errorf("request error"))

	logrus.WithFields(logrus.Fields{
		"code": errApp.Status(),
	}).Error(errApp.Error())

	jsonErr, err := errApp.Marshal()
	if err != nil {
		logrus.Errorf("Marshaling Error interface failed: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshal response error"))
		return
	}
	w.WriteHeader(errApp.Status())
	if _, err := w.Write(jsonErr); err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error writing response"))
		return
	}
}

func RequestOk(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
