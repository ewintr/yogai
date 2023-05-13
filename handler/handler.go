package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Index(w http.ResponseWriter) {
	Message(w, http.StatusOK, "yogai index")
}

func Message(w http.ResponseWriter, status int, message string, details ...any) {
	w.WriteHeader(status)
	response := struct {
		Message string `json:"message"`
		Details []any  `json:"details,omitempty"`
	}{
		Message: message,
		Details: details,
	}
	body, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		fmt.Fprintf(w, fmt.Sprintf(`{"message": %q, "details":%q}`, message, marshalErr.Error()))
		return
	}
	fmt.Fprintf(w, string(body))
}

func Error(w http.ResponseWriter, status int, message string, err error, details ...any) {
	w.WriteHeader(status)
	response := struct {
		Message string `json:"message"`
		Error   string `json:"error"`
		Details []any  `json:"details,omitempty"`
	}{
		Message: message,
		Error:   err.Error(),
		Details: details,
	}
	body, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		fmt.Fprintf(w, fmt.Sprintf(`{"message": %q, "error": %q, "details":%q}`, message, err.Error(), marshalErr.Error()))
		return
	}

	fmt.Fprintf(w, string(body))
}
