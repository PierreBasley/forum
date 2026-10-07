package server

import (
	"bytes"
	"net/http"
	"strconv"
)

const (
	ERR_PATH = "./templates/errors/"
)

type ErrorDetails struct {
	Error error
	AllowedMethods []string
}

func renderError(w http.ResponseWriter, status int, errorStruct *ErrorDetails) {
	errorCode := strconv.Itoa(status)
	errorPage := ERR_PATH + errorCode + ".html"

	t, err := parseTemplates(errorPage)
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	var buffer bytes.Buffer

	if errorStruct == nil {
		err = t.Execute(&buffer, nil)
	} else {
		err = t.Execute(&buffer, *errorStruct)
	}
	if err != nil {
		http.Error(w, http.StatusText(status), status)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(buffer.Len()))
	w.WriteHeader(status)

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		return
	}
}
