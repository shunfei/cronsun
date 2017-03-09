package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"

	"sunteng/commons/log"
)

type BaseHandler struct {
	Handle func(w http.ResponseWriter, r *http.Request)
}

func (b BaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// handle all the error
		err_ := recover()
		if err_ == nil {
			return
		}

		var stack string
		var buf bytes.Buffer
		buf.Write(debug.Stack())
		stack = buf.String()

		outJSONWithCode(w, http.StatusInternalServerError, "Internal Server Error")

		log.Errorf("%v\n\n%s\n", err_, stack)
		return
	}()

	b.Handle(w, r)
}

func outJSONWithCode(w http.ResponseWriter, httpCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	s := ""
	b, err := json.Marshal(data)
	if err != nil {
		s = `{"error":"json.Marshal error"}`
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		s = string(b)
		w.WriteHeader(httpCode)
	}
	fmt.Fprint(w, s)
}

func outJSON(w http.ResponseWriter, data interface{}) {
	outJSONWithCode(w, http.StatusOK, data)
}
