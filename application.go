package mango

import (
    "net/http"
    "fmt"
    "runtime"
    "os"
    "os/signal"
)

var (
    Debug = false
    CookieSecret = "fLjUfxqXtfNoIldA0A0J"
)

type Application struct {
    Addr string
    Port int
    Url map[string]interface{}
    StaticPath string
    TemplatePath string
}

func (app *Application) Run() {
    runtime.GOMAXPROCS(runtime.NumCPU())
    http.Handle("/",NewRouter(app.Url))
    if (app.StaticPath != "") {
        http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(app.StaticPath))))
    }
    if (app.TemplatePath != "") {
        TemplateLoader(app.TemplatePath)
    }

    fmt.Printf("http://%s:%d\n", app.Addr, app.Port)

    root := ""
    if cwd, err := os.Getwd(); err != nil {
        panic(fmt.Sprintf("Error getting working directory: %s", err))
    } else {
        root = cwd
    }

    if Debug {
        start(app.Addr, app.Port, root)
    } else {
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
}
