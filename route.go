package mango

import (
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"
)

type H map[string]interface{}

type Router struct {
	// Prepare
	PreHandler Handler
	// 404
	NotFoundHandler Handler
	// 500
	ErrorHandler Handler
	// See Router.StrictSlash(). This defines the flag for new routes.
	strictSlash bool
	//routers     map[*regexp.Regexp]interface{}
	routers map[*regexp.Regexp]map[string]Handler
}

// NewRouter returns a new router instance.
func NewRouter(urls map[string]map[string]Handler, n Handler, e Handler, p Handler) *Router {
	route := make(map[*regexp.Regexp]map[string]Handler)

	for k, _ := range urls {
		re, err := regexp.Compile("^" + k + "$")
		if err != nil {
			fmt.Println("url is error:", k)
		}
		route[re] = urls[k]
	}
	return &Router{routers: route, NotFoundHandler: n, ErrorHandler: e, PreHandler: p}
}

func (self *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Clean path to canonical form and redirect.
	var ctx *HTTPRequest
	start := time.Now()
	var st, l int
	defer func() {
		if err := recover(); err != nil {
			st = http.StatusInternalServerError
			var e = fmt.Sprintf("%s", err)
			fmt.Println(e)
			if self.ErrorHandler == nil {
				w.WriteHeader(http.StatusInternalServerError)
				if Debug {
					w.Write([]byte(e))
				} else {
					w.Write([]byte("Internal Server Error"))
				}
			} else {
				self.ErrorHandler(ctx)
			}
		}
		use := float64(time.Now().UnixNano() - start.UnixNano())

		var ip string
		if ctx != nil {
			st = ctx.StatusCode
			l = ctx.Length
			ip = ctx.RemoteAddr
		} else {
			if st == 0 {
				st = 200
			}
			l = 0
			ip = strings.Split(req.RemoteAddr, ":")[0]
		}

		fmt.Printf("%s - - [%d-%02d-%02d %02d:%02d:%02d] \"%s %s \" %d %d %.6f\n",
			ip, start.Year(), start.Month(), start.Day(),
			start.Hour(), start.Minute(), start.Second(),
			req.Method, req.URL.Path, st, l, use/1000000)
	}()

	if p := cleanPath(req.URL.Path); p != req.URL.Path {
		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	flag := false
	for k, _ := range self.routers {
		m := k.FindStringSubmatch(req.URL.Path)

		if len(m) != 0 {
			ctx = &HTTPRequest{}
			ctx.Init(w, req, m)
			if self.PreHandler != nil {
				self.PreHandler(ctx)
			}

			if f, ok := self.routers[k][req.Method]; ok {
				f(ctx)
			} else {
				http.Error(w, "Method Not Allowed", 405)
			}
			return
		}
	}
	if !flag {
		st = http.StatusNotFound
		if self.NotFoundHandler == nil {
			http.NotFoundHandler().ServeHTTP(w, req)
		} else {
			ctx := &HTTPRequest{}
			ctx.Init(w, req, []string{})
			self.NotFoundHandler(ctx)
		}
	}
}

func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}
