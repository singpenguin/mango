package web

import (
	"net/http"
	"fmt"
	"runtime"
	"os"
	"os/signal"
)

var (
	Debug = false
)

type Application struct {
	Addr string
	Port int
	Url map[string]interface{}
}

func (app *Application) Run() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.Handle("/",NewRouter(app.Url))

	fmt.Printf("http://%s:%d\n", app.Addr, app.Port)

	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", app.Addr, app.Port), nil)
		if err != nil {
			fmt.Println("Failed to start server:", err)
		}
	}()
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	<-ch
}
