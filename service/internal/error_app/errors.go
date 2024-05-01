package errorsApp

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// общие ошибки
	ErrEmptyField         = errors.New("задано пустое поле")
	ErrInvalidData        = errors.New("некорректное поле")
	ErrInvalidQueryParams = errors.New("некорректное значение параметра")
)

// Структура ошибок HTTP сервера
type AppError struct {
	Code   int    `json:"code"`  // код ошибки
	MsgErr string `json:"error"` // ошибка
}

// метод, реализующий интерфейс error
func (e *AppError) Error() string {
	return e.MsgErr
}

// метод, сереализующий структуру ошибки в формат JSON
func (e *AppError) Marshal() ([]byte, error) {
	jsonData, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

// метод, добавляющий префикс ошибки
func (e *AppError) WrapErr(prefix string) {
	e.MsgErr = strings.Join([]string{prefix, e.MsgErr}, ": ")
}

// конструктор структуры appError
func NewError(code int, err error) *AppError {
	return &AppError{Code: code, MsgErr: err.Error()}
}

// функция для формирования http заголовков для ответа клиенту
// в случае запроса, приведшего к ошибке
func RequestError(w http.ResponseWriter, typeRequest string, appErr *AppError) {
	w.Header().Set("Content-Type", "application/json")
	// Добавление к ошибке префикса
	appErr.WrapErr("request error")

	logrus.WithFields(logrus.Fields{
		"method": typeRequest,
		"code":   appErr.Code,
	}).Error(appErr.Error())

	// сереализация ошибки
	jsonErr, err := appErr.Marshal()
	if err != nil {
		logrus.Errorf("Marshaling Error interface failed: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshal response error"))
		return
	}
	w.WriteHeader(appErr.Code)
	if _, err := w.Write(jsonErr); err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error writing response"))
		return
	}
}

// функция для формирования http заголовков и ответа клиенту
// в случае успешного запроса
func RequestOk(w http.ResponseWriter, typeRequest string, bodyResponse io.Reader) {
	// запись заголовков
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// чтение ответа из io.Reader
	data, err := io.ReadAll(bodyResponse)
	if err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
		return
	}
	// ответ клиенту
	if _, err := w.Write(data); err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error writing response"))
		return
	}

	logrus.WithFields(logrus.Fields{
		"method": typeRequest,
	}).Info("request to service")
}
