package controllers

import (
	"github.com/singpenguin/mango"
)

type Index struct {
	mango.HTTPRequest
}
func (self *Index) GET() {
	//set header
	self.SetHeader("X-File", "true")
	//set cookie	key value max-age(second)
	self.SetCookie("username", "mango", 30)
	//output template
	self.Render("index.html", nil)
}

type News struct {
	mango.HTTPRequest
}

func (self *News) GET() {
	//output string
	self.Write("news")
}
