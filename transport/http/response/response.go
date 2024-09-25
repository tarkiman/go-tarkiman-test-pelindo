package response

import (
	"encoding/json"
	"net/http"

	"github.com/tarkiman/go/shared/failure"
	"github.com/tarkiman/go/shared/logger"
)

// Base is the base object of all responses
type Base struct {
	Data    *interface{} `json:"data,omitempty"`
	Error   *string      `json:"error,omitempty"`
	Message *string      `json:"message,omitempty"`
}

// NoContent sends a response without any content
func NoContent(w http.ResponseWriter) {
	respond(w, http.StatusNoContent, nil)
}

// WithMessage sends a response with a simple text message
func WithMessage(w http.ResponseWriter, code int, message string) {
	respond(w, code, Base{Message: &message})
}

// WithJson sends a response containing a JSON object
func WithJson(w http.ResponseWriter, code int, jsonPayload interface{}) {
	respond(w, code, Base{Data: &jsonPayload})
}

// WithJSON sends a response containing a JSON object
func WithJSON(w http.ResponseWriter, code int, jsonPayload interface{}) {
	respond(w, code, &jsonPayload)
}

// WithError sends a response with an error message
func WithError(w http.ResponseWriter, err error) {
	code := failure.GetCode(err)
	errMsg := err.Error()
	respond(w, code, Base{Error: &errMsg})
}

// WithPreparingShutdown sends a default response for when the server is preparing to shut down
func WithPreparingShutdown(w http.ResponseWriter) {
	WithMessage(w, http.StatusServiceUnavailable, "SERVER PREPARING TO SHUT DOWN")
}

// WithUnhealthy sends a default response for when the server is unhealthy
func WithUnhealthy(w http.ResponseWriter) {
	WithMessage(w, http.StatusServiceUnavailable, "SERVER UNHEALTHY")
}

func respond(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err := w.Write(response)
	if err != nil {
		logger.ErrorWithStack(err)
	}
}

// Auth Response
type Legacy struct {
	Name    string      `json:"name,omitempty"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Status  int         `json:"status"`
	Params  interface{} `json:"params,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func RenderLegacyError(w http.ResponseWriter, httpCode int, err error) {
	respond(w, httpCode, Legacy{Name: http.StatusText(httpCode), Message: failure.GetMessage(err), Code: 0, Status: httpCode})
}

func RenderLegacyFailure(w http.ResponseWriter, err error) {
	httpCode := failure.GetCode(err)
	respond(w, httpCode, Legacy{Name: http.StatusText(httpCode), Message: failure.GetMessage(err), Code: 0, Status: httpCode, Data: failure.GetData(err)})
}

func RenderLegacyAuthSuccess(w http.ResponseWriter, httpCode int, data interface{}) {
	respond(w, httpCode, data)
}

func RenderSuccess(w http.ResponseWriter, code int, message string, params interface{}, data interface{}) {
	pr := params
	if pr == nil {
		pr = make([]interface{}, 0)
	}

	respond(w, code, Base{
		Message: &message,
		Data:    &data,
	})
}
