package server

import (
	"bytes"
	"html/template"
	"net/http"
	"strconv"
)

func parseTemplates(filenames string) (*template.Template, error) {
	path := "./templates/layouts/"
	t, err := template.ParseFiles(
		filenames,
		path+"head.html",
		path+"header.html",
		path+"footer.html")

	return t, err
}

func executeBuffer(t *template.Template, w http.ResponseWriter) (http.ResponseWriter, error) {
	var buffer bytes.Buffer

	err := t.Execute(&buffer, nil) //add the right variable as data to be rendered in the website
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return w, err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(buffer.Len())) //Eventually, AI told me this part is not an obligation because the go server can handle this without problem

	_, err = w.Write(buffer.Bytes())
	if err != nil {
		return w, err
	}
	return w, err
}
