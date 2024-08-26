package controllers

import "net/http"

func NotFound(w http.ResponseWriter, r *http.Request) {
	code := http.StatusNotFound
	w.WriteHeader(code)
	w.Write(responseErrorBytes(code))
}
