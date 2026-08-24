package httpresponse

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Error   *ErrorBody  `json:"error"`
}

type ErrorBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Success(w http.ResponseWriter, status int, data interface{}) {
	writeJSON(w, status, Response{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

func Error(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, Response{
		Success: false,
		Data:    nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

func ErrorWithDetails(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
	details interface{},
) {
	writeJSON(w, status, Response{
		Success: false,
		Data:    nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func writeJSON(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(response)
}

// SuccessGin envía una respuesta JSON estructurada de éxito para Gin
func SuccessGin(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
		Error:   nil,
	})
}

// ErrorGin envía una respuesta JSON estructurada de error para Gin
func ErrorGin(c *gin.Context, status int, code string, message string) {
	c.JSON(status, Response{
		Success: false,
		Data:    nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
		},
	})
}

// ErrorWithDetailsGin envía una respuesta JSON estructurada de error con detalles adicionales para Gin
func ErrorWithDetailsGin(c *gin.Context, status int, code string, message string, details interface{}) {
	c.JSON(status, Response{
		Success: false,
		Data:    nil,
		Error: &ErrorBody{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}
