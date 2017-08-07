package main

import (
	"net/http"
)

func handleRunPython(w http.ResponseWriter, r *http.Request) {
		if !checkLogin(r) {
			do403(w)
			return
		}
    url:= "http://localhost:8000/"
    r.Close = true
    r.Header.Add("Accept-Encoding", "identity")
    http.Redirect(w, r, url, 307)
}
