package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"
)

type WithStatus interface {
	Status() int
}

type OkStatus struct{}

func (s *OkStatus) Status() int {
	return 200
}

type JsonRequestHandler func(req *http.Request) WithStatus
type JsonBodyRequestHandler func(req *http.Request, body interface{}) WithStatus

type JsonHandler struct {
	typ reflect.Type
	fun reflect.Value
}

func NewJsonHandler(i interface{}) http.Handler {
	t := reflect.TypeOf(i)
	if reflect.Func != t.Kind() {
		panic("JSON handler: must be a func")
	}
	if t.NumOut() != 1 {
		panic("JSON handler: must return a single value")
	}
	if t.NumIn() != 1 && t.NumIn() != 2 {
		panic("JSON handler: arity must be 1 or 2")
	}
	if t.In(0).String() != "*http.Request" {
		panic("JSON handler: first argument must be a *http.Request, was " + t.In(0).String())
	}
	if t.NumIn() == 2 && t.In(1).Kind() != reflect.Ptr {
		panic("JSON handler: second argument must be a pointer type, was " + t.In(1).String())
	}
	if t.Out(0).String() != "api.WithStatus" {
		panic("JSON handler: return value must implement WithStatus, was " + t.Out(0).String())
	}
	var jsonType reflect.Type
	if t.NumIn() == 2 {
		jsonType = t.In(1).Elem()
	}
	return JsonHandler{
		typ: jsonType,
		fun: reflect.ValueOf(i),
	}
}

func (h JsonHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	var returnValue reflect.Value
	if h.typ == nil {
		returnValue = h.fun.Call([]reflect.Value{
			reflect.ValueOf(req),
		})[0]
	} else {
		body := reflect.New(h.typ)
		json.NewDecoder(req.Body).Decode(body.Interface())
		returnValue = h.fun.Call([]reflect.Value{
			reflect.ValueOf(req),
			body,
		})[0]
	}
	withStatus, ok := returnValue.Interface().(WithStatus)
	if !ok {
		log.Println("failed to cast to WithStatus: ", returnValue)
		writeJsonResponse(rw, 500, map[string]string{
			"errcode": "M_SERVER_ERROR",
			"error":   "failed to read response status",
		})
	} else {
		writeJsonResponseWithStatus(rw, withStatus)
	}
}

func writeJsonResponse(rw http.ResponseWriter, status int, body interface{}) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")
	res, err := json.Marshal(body)
	if err != nil {
		rw.WriteHeader(500)
		log.Println("marshaling error: ", err)
		fmt.Fprintf(rw, "{\"errcode\":\"M_SERVER_ERROR\",\"error\":\"failed to marshal response\"}")
	} else {
		rw.WriteHeader(status)
		rw.Write(res)
	}
}

func writeJsonResponseWithStatus(rw http.ResponseWriter, body WithStatus) {
	writeJsonResponse(rw, body.Status(), body)
}

type Resource struct {
	Get    http.Handler
	Post   http.Handler
	Put    http.Handler
	Delete http.Handler
}

func (r Resource) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	log.Println("req: " + req.Method)
	switch req.Method {
	case "GET":
		if r.Get != nil {
			r.Get.ServeHTTP(rw, req)
			return
		}
	case "POST":
		if r.Post != nil {
			r.Post.ServeHTTP(rw, req)
			return
		}
	case "PUT":
		if r.Put != nil {
			r.Put.ServeHTTP(rw, req)
			return
		}
	case "DELETE":
		if r.Delete != nil {
			r.Delete.ServeHTTP(rw, req)
			return
		}
	}
	writeJsonResponseWithStatus(rw, defaultUnrecognizedError)
}
