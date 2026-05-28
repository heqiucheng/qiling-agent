package httpx

import (
	"encoding/json"
	"net/http"
	"time"
)

type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type ErrorBody struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

type Response struct {
	Data  any        `json:"data"`
	Error *ErrorBody `json:"error"`
	Meta  Meta       `json:"meta"`
}

func WriteJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	write(w, status, Response{Data: data, Error: nil, Meta: metaFromRequest(r)})
}

func WriteError(w http.ResponseWriter, r *http.Request, status int, code string, message string, details map[string]any) {
	write(w, status, Response{
		Data:  nil,
		Error: &ErrorBody{Code: code, Message: message, Details: details},
		Meta:  metaFromRequest(r),
	})
}

func write(w http.ResponseWriter, status int, response Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}

func metaFromRequest(r *http.Request) Meta {
	requestID, _ := r.Context().Value(requestIDKey{}).(string)
	return Meta{RequestID: requestID, Timestamp: time.Now().UTC().Format(time.RFC3339)}
}
