package api

import "net/http"

var Root *http.ServeMux

func init() {
	Root = http.NewServeMux()
	Root.HandleFunc("/", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("ROOT"))
	})
	Root.HandleFunc("/test", func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte("TEST"))
	})
}
