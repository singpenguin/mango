package web

import (
	"net/http"
	"fmt"
	"strings"
	"io/ioutil"
	"bytes"
	"mime/multipart"
	"net/url"
)

type HTTPRequest struct {
	Args []string
	Method string
	RemoteAddr string
	Path string
	Host string
	Params url.Values
	Body []byte
	Files map[string][]*multipart.FileHeader

	W http.ResponseWriter
	R *http.Request
}

func (self *HTTPRequest) Init(w http.ResponseWriter, r *http.Request, a []string) {
	self.W = w
	self.R = r
	self.Method = r.Method
	self.Path = r.URL.Path
	self.Args = a
	self.Host = r.Host

	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	self.RemoteAddr = strings.Split(ip, ":")[0]
	self.parseBodyArguments()
}

func (self *HTTPRequest) UserAgent() string {
	return self.R.Header.Get("User-Agent")
}

func (self *HTTPRequest) Header(s string) string {
	return self.R.Header.Get(s)
}

func (self *HTTPRequest) SetHeader(k, v string) {
	self.W.Header().Set(k, v)
}

func (self *HTTPRequest) SetStatus(st int) {
	self.W.WriteHeader(st)
}

func (self *HTTPRequest) Cookie(k string) string {
	ck, err := self.R.Cookie(k)
	if err != nil {
		return ""
	}
	return ck.Value
}

var cks = strings.NewReplacer("\n", "-", "\r", "-")
var cvs = strings.NewReplacer("\n", " ", "\r", " ", ";", " ")

func (self *HTTPRequest) SetCookie(k, v string, params ...interface{}) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s=%s", cks.Replace(k), cvs.Replace(v))
	if len(params) > 0 {
		switch v := params[0].(type) {
		case int:
			if v > 0 {
				fmt.Fprintf(&b, "; Max-Age=%d", v)
			} else if v < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		case int64:
			if v > 0 {
				fmt.Fprintf(&b, "; Max-Age=%d", v)
			} else if v < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		case int32:
			if v > 0 {
				fmt.Fprintf(&b, "; Max-Age=%d", v)
			} else if v < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		}
	}
	// Path, Domain, Secure, HttpOnly

	// default "/"
	if len(params) > 1 {
		if v, ok := params[1].(string); ok && len(v) > 0 {
			fmt.Fprintf(&b, "; Path=%s", cvs.Replace(v))
		}
	} else {
		fmt.Fprintf(&b, "; Path=%s", "/")
	}

	// default empty
	if len(params) > 2 {
		if v, ok := params[2].(string); ok && len(v) > 0 {
			fmt.Fprintf(&b, "; Domain=%s", cvs.Replace(v))
		}
	}

	// default empty
	if len(params) > 3 {
		var secure bool
		switch v := params[3].(type) {
		case bool:
			secure = v
		default:
			if params[3] != nil {
				secure = true
			}
		}
		if secure {
			fmt.Fprintf(&b, "; Secure")
		}
	}

	// default false. 
	httponly := false
	if len(params) > 4 {
		if v, ok := params[4].(bool); ok && v {
			// HttpOnly = true
			httponly = true
		}
	}

	if httponly {
		fmt.Fprintf(&b, "; HttpOnly")
	}

	self.W.Header().Add("Set-Cookie", b.String())
}

func (self *HTTPRequest) Scheme() string {
	if self.R.URL.Scheme != "" {
		return self.R.URL.Scheme
	} else if self.R.TLS == nil {
		return "http"
	} else {
		return "https"
	}
}

func (self *HTTPRequest) Redirect(newloc string) {
	self.SetHeader("Location", newloc)
	self.SetStatus(303)
}

func (self *HTTPRequest) Query() map[string][]string {
	return self.R.URL.Query()
}

func (self *HTTPRequest) QueryString() string {
	if s := strings.Split(self.R.RequestURI, "?"); len(s) > 1 {
		return s[1]
	}else{
		return ""
	}
}

func (self *HTTPRequest) parseBodyArguments() {
	qs := self.R.URL.Query()
	var pt url.Values
	ct := self.Header("Content-Type")
	if strings.HasPrefix(ct, "application/x-www-form-urlencoded") {
		if err := self.R.ParseForm(); err != nil {
			fmt.Println("Error parsing request body:", err)
		} else {
			pt = self.R.Form
		}
	} else if strings.HasPrefix(ct, "multipart/form-data") {
		// TODO: Extract the multipart form param so app can set it.
		if err := self.R.ParseMultipartForm(32 << 20 /* 32 MB */); err != nil {
			fmt.Println("Error parsing request body:", err)
		} else {
			self.Files = self.R.MultipartForm.File
		}
	} else if strings.HasPrefix(ct, "application/octet-stream") {
		self.Body = self.Data()
	}

	num := len(qs) + len(pt)

	if num == 0 {
		self.Params = make(url.Values, 0)
		return
	}
	values := make(url.Values, num)
	for k, v := range qs {
		values[k] = append(values[k], v...)
	}
	for k, v := range pt {
		values[k] = append(values[k], v...)
	}
	self.Params = values
}

func (self *HTTPRequest) Data() []byte {
	body, _ := ioutil.ReadAll(self.R.Body)
	self.R.Body.Close()
	bf := bytes.NewBuffer(body)
	self.R.Body = ioutil.NopCloser(bf)
	return body
}

func (self *HTTPRequest) Write(s string) {
	self.W.Write([]byte(s))
}

func (self *HTTPRequest) GET() {
	http.Error(self.W, "Method Not Allowed", 405)
}

func (self *HTTPRequest) POST() {
	http.Error(self.W, "Method Not Allowed", 405)
}

func (self *HTTPRequest) HEAD() {
	http.Error(self.W, "Method Not Allowed", 405)
}

func (self *HTTPRequest) OPTION() {
	http.Error(self.W, "Method Not Allowed", 405)
}

func (self *HTTPRequest) PUT() {
	http.Error(self.W, "Method Not Allowed", 405)
}
