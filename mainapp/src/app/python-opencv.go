package main

import (
	"net/http"
	"log"
	"bytes"
	"net/url"
)

func handleRunPython(w http.ResponseWriter, r *http.Request) {
		if !checkLogin(r) {
			do403(w)
			return
		}
		url, err := url.Parse("http://compile-service:8000/")
		log.Println(url.String())
		if err != nil {
			log.Println(err)
		}
		proxyReq, err := http.NewRequest(r.Method, url.String(), r.Body)
		if err != nil {
			log.Println("request error: ", err.Error())

		}

		proxyReq.Header = r.Header

		client := &http.Client{}
		proxyRes, err := client.Do(proxyReq)
		if err != nil {
			log.Println("request to compile error: ", err.Error())
		}
		buf := bytes.NewBuffer(make([]byte, 0, proxyRes.ContentLength))
		_, readErr := buf.ReadFrom(proxyRes.Body)
		if readErr != nil {
			log.Println("ERROR READING RESPONSE: ", readErr.Error())
		}
		body := buf.Bytes()

		w.Header().Set("Content-Type", "application/json")
  	w.Write(body)
}
