package mango

import (
	"net/http"
	"path"
	"regexp"
	"fmt"
	"reflect"
)

type Router struct {
	// 自定义的404
	NotFoundHandler http.HandlerFunc
	// See Router.StrictSlash(). This defines the flag for new routes.
	strictSlash bool
	routers map[*regexp.Regexp]interface{}
}

// NewRouter returns a new router instance.
func NewRouter(urls map[string]interface{}) *Router {
	route := make(map[*regexp.Regexp]interface{})

	for k,_ := range urls {
		re, err := regexp.Compile("^" + k + "$")
		if err != nil {
			fmt.Println("url is error:", k)
		}
		route[re] = urls[k]
	}
	return &Router{routers:route}
}

func (self *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Clean path to canonical form and redirect.
	if p := cleanPath(req.URL.Path); p != req.URL.Path {
		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently)
		return
	}
	flag := false
	for k,_ := range self.routers {
		m := k.FindStringSubmatch(req.URL.Path)

		if len(m) !=0 {
			t := reflect.TypeOf(self.routers[k]).Elem()
			v := reflect.New(t)
			in := []reflect.Value{reflect.ValueOf(w), reflect.ValueOf(req), reflect.ValueOf(m)}
			v.MethodByName("Init").Call(in)
			v.MethodByName(req.Method).Call(nil)

			return
		}
	}
	if flag == false {
		if self.NotFoundHandler == nil {
			http.NotFoundHandler().ServeHTTP(w,req)
		}
		//self.NotFoundHandler(w, req)
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
