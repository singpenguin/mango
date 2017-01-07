package mango

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/singpenguin/mango/utils"
)

type HTTPRequest struct {
	Args       []string //route params
	Method     string   //HTTP method
	RemoteAddr string   //ip
	Path       string
	Host       string
	Params     url.Values                         //GET or POST params
	Body       []byte                             //Content-Type=application/octet-stream, Body=http body
	Files      map[string][]*multipart.FileHeader //Content-Type=multipart/form-data
	Length     int                                //response length
	StatusCode int

	W http.ResponseWriter
	R *http.Request
}

func (self *HTTPRequest) Init(w http.ResponseWriter, r *http.Request, a []string) {
	self.W = w
	self.R = r
	//self.Method = r.Method
	self.Path = r.URL.Path
	self.Args = a
	self.Host = r.Host
	self.StatusCode = http.StatusOK

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
	self.StatusCode = st
	self.W.WriteHeader(st)
}

func (self *HTTPRequest) Cookie(k string) (string, error) {
	ck, err := self.R.Cookie(k)
	if err != nil {
		return "", err
	}
	return ck.Value, nil
}

var cks = strings.NewReplacer("\n", "-", "\r", "-")
var cvs = strings.NewReplacer("\n", " ", "\r", " ", ";", " ")

func (self *HTTPRequest) SetSecureCookie(k, v string, params ...interface{}) {
	bv := []byte(v)
	ts := utils.IntToStr(utils.Timestamp())
	v = utils.Base64Encode(bv)
	sa := []string{}
	sa = append(sa, "2|1:0")
	sa = append(sa, fmt.Sprintf("%d:%s", len(ts), ts))
	sa = append(sa, fmt.Sprintf("%d:%s", len(k), k))
	sa = append(sa, fmt.Sprintf("%d:%s", len(v), v))
	sa = append(sa, "")
	to_sign := strings.Join(sa, "|")
	signature := utils.HmacSha256(to_sign, CookieSecret)

	self.SetCookie(k, to_sign+signature, params...)
}

func (self *HTTPRequest) GetSecureCookie(k string) (string, error) {
	value, err := self.Cookie(k)
	if err != nil || len(value) < 64 {
		return "", err
	}
	_rest := value[2:]
	_, rest, err := consumeField(_rest)
	if err != nil {
		return "", err
	}
	timestamp_field, rest, err := consumeField(rest)
	if err != nil {
		return "", err
	}
	name_field, rest, err := consumeField(rest)
	if err != nil {
		return "", err
	}
	value_field, rest, err := consumeField(rest)
	if err != nil {
		return "", err
	}
	passed_sig := rest
	signed_str := value[:len(value)-len(passed_sig)]
	signature := utils.HmacSha256(signed_str, CookieSecret)
	if passed_sig != signature {
		return "", errors.New("signature invalid")
	}
	if name_field != k {
		return "", errors.New("cookie name invalid")
	}
	timestamp, _ := utils.StrToInt64(timestamp_field)
	if timestamp < utils.Timestamp()-2678400 {
		return "", errors.New("signature has expired")
	}
	devalue, err := utils.Base64Decode(value_field)
	if err != nil {
		return "", errors.New("value cannot be decode")
	}
	return devalue, nil
}

func consumeField(value string) (string, string, error) {
	var length, rest string
	utils.Unpack(strings.SplitN(value, ":", 2), &length, &rest)

	n, err := utils.StrToInt64(length)
	if err != nil {
		return "", "", errors.New("length field invalid")
	}
	key_field := rest[:n]
	if rest[n:n+1] != "|" {
		return "", "", errors.New("malformed signed value")
	}
	rest = rest[n+1:]
	return key_field, rest, nil
}

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
	self.StatusCode = http.StatusFound
	self.SetHeader("Location", newloc)
	self.SetStatus(http.StatusFound)
}

func (self *HTTPRequest) Query() map[string][]string {
	return self.R.URL.Query()
}

func (self *HTTPRequest) QueryString() string {
	if s := strings.Split(self.R.RequestURI, "?"); len(s) > 1 {
		return s[1]
	} else {
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

func (self *HTTPRequest) Write(s []byte) {
	n, _ := self.W.Write(s)
	self.Length += n
}

func (self *HTTPRequest) Render(name string, data interface{}) {
	Template[name].Execute(self.W, data)
}

var jsonContentType = "application/json; charset=utf-8"

func (self *HTTPRequest) Send(data interface{}) error {
	self.SetHeader("Content-Type", jsonContentType)
	return json.NewEncoder(self.W).Encode(data)
}

type Handler func(ctx *HTTPRequest)
