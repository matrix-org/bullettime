package api

import (
	"log"
	"net/http"
)

var Root *http.ServeMux

type TestRequest struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

type TestResponse struct {
	OkStatus
	Message string `json:"message"`
}

func handlePost(req *http.Request, body *TestRequest) WithStatus {
	log.Println("got body", body.Code, body.Message)
	return &TestResponse{
		Message: "hello",
	}
}

func init() {
	Root = http.NewServeMux()
	Root.Handle("/", NewJsonHandler(func(req *http.Request) WithStatus {
		return defaultUnrecognizedError
	}))
	Root.Handle("/test", NewJsonHandler(handlePost))
}
