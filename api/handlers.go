package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"reflect"

	"github.com/Rugvip/bullettime/service"

	"github.com/julienschmidt/httprouter"
)

type WithStatus interface {
	Status() int
}

type OkStatus struct{}

func (s *OkStatus) Status() int {
	return 200
}

type JsonHandler func(req *http.Request, params httprouter.Params) interface{}
type JsonBodyHandler func(req *http.Request, params httprouter.Params, body interface{}) interface{}

func jsonHandler(i interface{}) httprouter.Handle {
	t := reflect.TypeOf(i)
	if reflect.Func != t.Kind() {
		panic("jsonHandler: must be a function")
	}
	if t.NumOut() != 1 {
		panic("jsonHandler: must return a single value")
	}
	argCount := t.NumIn()

	var jsonType reflect.Type
	firstParamIsParams := false
	if argCount > 0 {
		firstParamIsParams = t.In(0).String() == "httprouter.Params"
		lastParamIsParams := t.In(argCount-1).String() == "httprouter.Params"
		lastParamIsRequest := t.In(argCount-1).String() == "*http.Request"
		if !lastParamIsParams && !lastParamIsRequest {
			if t.In(argCount-1).Kind() != reflect.Ptr {
				panic("jsonHandler: body argument must be a pointer type, was " + t.In(2).String())
			}
			jsonType = t.In(argCount - 1).Elem()
		}
	}
	if jsonType == nil {
		if t.NumIn() > 2 {
			panic("jsonHandler: arity must be at most 2 if no body argument is preset")
		}
	} else {
		if t.NumIn() > 3 {
			panic("jsonHandler: arity must be at most 3 if body argument is preset")
		}
	}
	handlerFunc := reflect.ValueOf(i)

	return func(rw http.ResponseWriter, req *http.Request, params httprouter.Params) {
		var returnValue reflect.Value
		var args []reflect.Value
		if jsonType == nil {
			switch argCount {
			case 0:
				args = []reflect.Value{}
			case 1:
				if firstParamIsParams {
					args = []reflect.Value{reflect.ValueOf(params)}
				} else {
					args = []reflect.Value{reflect.ValueOf(req)}
				}
			case 2:
				args = []reflect.Value{reflect.ValueOf(req), reflect.ValueOf(params)}
			}
		} else {
			body := reflect.New(jsonType)
			if err := json.NewDecoder(req.Body).Decode(body.Interface()); err != nil {
				writeJsonResponseWithStatus(rw, BadJsonError(err.Error()))
				return
			}
			switch argCount {
			case 1:
				args = []reflect.Value{body}
			case 2:
				if firstParamIsParams {
					args = []reflect.Value{reflect.ValueOf(params), body}
				} else {
					args = []reflect.Value{reflect.ValueOf(req), body}
				}
			case 3:
				args = []reflect.Value{reflect.ValueOf(req), reflect.ValueOf(params), body}
			}
		}
		returnValue = handlerFunc.Call(args)[0]
		res := returnValue.Interface()

		withStatus, ok := res.(WithStatus)
		if ok {
			writeJsonResponseWithStatus(rw, withStatus)
		} else {
			writeJsonResponse(rw, 200, res)
		}
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

func readAccessToken(req *http.Request) (service.User, *Error) {
	token := req.URL.Query().Get("access_token")
	if token == "" {
		return service.User{}, defaultMissingTokenError
	}
	info, err := service.ParseAccessToken(token)
	if err != nil {
		return service.User{}, defaultUnknownTokenError
	}
	user, err := service.GetUser(info.UserId)
	if err != nil {
		return service.User{}, defaultUnknownTokenError
	}
	return user, nil
}
