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

	"github.com/shunfei/cronsun"
	"github.com/shunfei/cronsun/conf"
	"github.com/shunfei/cronsun/log"
	"github.com/shunfei/cronsun/web/session"
)

var sessManager session.SessionManager

func InitServer() (*http.Server, error) {
	sessManager = session.NewEtcdStore(cronsun.DefalutClient, conf.Config.Web.Session)

	var err error
	if err = checkAuthBasicData(); err != nil {
		return nil, err
	}

	return initRouters()
}

type Context struct {
	Data    map[interface{}]interface{}
	Session *session.Session
	todos   []func()
	R       *http.Request
	W       http.ResponseWriter
}

func newContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		R:     r,
		W:     w,
		todos: make([]func(), 0, 2),
		Data:  make(map[interface{}]interface{}, 2),
	}
}

func (ctx *Context) Todo(f func()) {
	ctx.todos = append(ctx.todos, f)
}

func (ctx *Context) Done() {
	n := len(ctx.todos) - 1
	for i := n; i >= 0; i-- {
		ctx.todos[i]()
	}
}

type BaseHandler struct {
	Ctx          map[string]interface{}
	BeforeHandle func(ctx *Context) (abort bool)
	Handle       func(ctx *Context)
}

func NewBaseHandler(f func(ctx *Context)) BaseHandler {
	return BaseHandler{
		BeforeHandle: authHandler(false),
		Handle:       f,
	}
}

func NewAuthHandler(f func(ctx *Context)) BaseHandler {
	return BaseHandler{
		BeforeHandle: authHandler(true),
		Handle:       f,
	}
}

func authHandler(needAuth bool) func(*Context) bool {
	return func(ctx *Context) (abort bool) {
		var err error
		ctx.Session, err = sessManager.Get(ctx.W, ctx.R)
		if ctx.Session != nil {
			ctx.Todo(func() {
				if err := sessManager.Store(ctx.Session); err != nil {
					log.Errorf("Failed to store session: %s.", err.Error())
				}
			})
		}

		if err != nil {
			outJSONWithCode(ctx.W, http.StatusInternalServerError, err.Error())
			abort = true
			return
		}

		if !conf.Config.Web.Auth.Enabled || !needAuth {
			return
		}

		if len(ctx.Session.Email) < 1 {
			outJSONWithCode(ctx.W, http.StatusUnauthorized, "please login.")
			abort = true
			return
		}
		return
	}
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

	ctx := newContext(w, r)
	defer ctx.Done()

	if b.BeforeHandle != nil {
		abort := b.BeforeHandle(ctx)
		if abort {
			return
		}
	}
	b.Handle(ctx)
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
