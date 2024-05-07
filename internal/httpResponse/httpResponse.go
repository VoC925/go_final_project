package httpResponse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	// общие ошибки
	ErrEmptyField         = errors.New("задано пустое поле")
	ErrInvalidData        = errors.New("некорректное поле")
	ErrInvalidQueryParams = errors.New("некорректное значение параметра")
	ErrNoData             = errors.New("данные отсутствуют")
	ErrUnAuth             = errors.New("ошибка аутентификации")
)

// Структура ошибок HTTP сервера
type AppError struct {
	Code   string `json:"code"`  // код ошибки
	MsgErr string `json:"error"` // сообщение ошибки
}

// метод, реализующий интерфейс error
func (e *AppError) Error() string {
	return e.MsgErr
}

// метод, добавляющий префикс ошибки
func (e *AppError) WrapErr(prefix string) {
	e.MsgErr = strings.Join([]string{prefix, e.MsgErr}, ": ")
}

// конструктор структуры appError
func NewError(code int, err error) *AppError {
	return &AppError{Code: fmt.Sprint(code), MsgErr: err.Error()}
}

// структура для формирования логов
type logInfo struct {
	cid           string        // id лога
	r             *http.Request // запрос
	bodyResoponse []byte        // тело ответа
	timeResponse  time.Duration // время ответа
	err           *AppError     // возможная ошибка
}

// конструктор структуры логов logInfo
func NewLogInfo(cid string, r *http.Request, body []byte, time time.Duration, appErr *AppError) *logInfo {
	return &logInfo{
		cid:           cid,
		r:             r,
		bodyResoponse: body,
		timeResponse:  time,
		err:           appErr,
	}
}

type msgErr struct {
	Msg string `json:"error"`
}

// функция для формирования http заголовков для ответа клиенту
// в случае запроса, приведшего к ошибке
func Error(w http.ResponseWriter, info *logInfo) {
	w.Header().Set("Content-Type", "application/json")
	// Добавление к ошибке префикса
	info.err.WrapErr("request error")

	// сереализация ошибки для клиента
	var msg msgErr
	// лсслучай когда ошибка - ошибка аутентификации
	if info.err.Code == "401" {
		msg.Msg = "Ошибка аутентификации"
	} else {
		msg.Msg = "Произошла ошибка при обработке вашего запроса"
	}
	// сереализация ошибки
	jsonErr, err := json.Marshal(msg)
	if err != nil {
		logrus.Errorf("Marshaling Error failed: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshal response error"))
		return
	}
	// запись заголовка
	code, _ := strconv.Atoi(info.err.Code)
	w.WriteHeader(code)
	// запись ответа
	if _, err := w.Write(jsonErr); err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error writing response"))
		return
	}
	// логирование
	logrus.WithFields(logrus.Fields{
		"log_ID":        info.cid,
		"URL":           info.r.URL.String(),
		"method_HTTP":   info.r.Method,
		"HTTP_Status":   info.err.Code,
		"size_response": len(jsonErr),
		"time_response": info.timeResponse.String(),
	}).Error(info.err.MsgErr)

}

// функция для формирования http заголовков и ответа клиенту
// в случае успешного запроса
func Success(w http.ResponseWriter, info *logInfo) {
	// запись заголовков
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// ответ клиенту
	if _, err := w.Write(info.bodyResoponse); err != nil {
		logrus.Errorf("Error writing response: %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error writing response"))
		return
	}
	// логирование
	logrus.WithFields(logrus.Fields{
		"log_ID":        info.cid,
		"URL":           info.r.URL.String(),
		"method_HTTP":   info.r.Method,
		"HTTP_Status":   http.StatusOK,
		"size_response": len(info.bodyResoponse),
		"time_response": info.timeResponse.String(),
	}).Info("request to service")
}
