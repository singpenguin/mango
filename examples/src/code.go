package main

import (
	"config"
	"github.com/singpenguin/mango"
)

func main(){
	app := &mango.Application{Addr:"127.0.0.1", Port:8888, Url:config.Urls, StaticPath:"static"}
	mango.TemplateLoader("templates")
	app.Run()
}
