package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/shunfei/cronsun/log"
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

func getStringArrayFromQuery(name, sep string, r *http.Request) (arr []string) {
	val := strings.TrimSpace(r.FormValue(name))
	if len(val) == 0 {
		return
	}

	return strings.Split(val, sep)
}

func getPage(page string) int {
	p, err := strconv.Atoi(page)
	if err != nil || p < 1 {
		p = 1
	}

	return p
}

func getPageSize(ps string) int {
	p, err := strconv.Atoi(ps)
	if err != nil || p < 1 {
		p = 50
	} else if p > 200 {
		p = 200
	}
	return p
}

func getTime(t string) time.Time {
	t = strings.TrimSpace(t)
	time, _ := time.ParseInLocation("2006-01-02", t, time.Local)
	return time
}

func getStringVal(n string, r *http.Request) string {
	return strings.TrimSpace(r.FormValue(n))
}

func InStringArray(k string, ss []string) bool {
	for i := range ss {
		if ss[i] == k {
			return true
		}
	}

	return false
}

func UniqueStringArray(a []string) []string {
	al := len(a)
	if al == 0 {
		return a
	}

	ret := make([]string, al)
	index := 0

loopa:
	for i := 0; i < al; i++ {
		for j := 0; j < index; j++ {
			if a[i] == ret[j] {
				continue loopa
			}
		}
		ret[index] = a[i]
		index++
	}

	return ret[:index]
}

// 返回存在于 a 且不存在于 b 中的元素集合
func SubtractStringArray(a, b []string) (c []string) {
	c = []string{}

	for _, _a := range a {
		if !InStringArray(_a, b) {
			c = append(c, _a)
		}
	}

	return
}
