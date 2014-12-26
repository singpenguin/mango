package mango

//
//set "mango.Debug = true" to turn on debug mode.
//Description:
//  Automatic restart in debug mode
//  Fork from https://github.com/fengsp/knight
//

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "time"
)

var mTimes = make(map[string]time.Time)

func visit(path string, info os.FileInfo, err error) error {
    if err != nil {
        log.Fatal(fmt.Sprintf("Error walking %s: ", path), err)
    }
    if !info.IsDir() && strings.HasSuffix(path, "go") {
        mTime := info.ModTime()
        oldTime, ok := mTimes[path]
        if ok {
            if oldTime.Before(mTime) {
                fmt.Println(" * Detected change, reloading")
                os.Exit(3)
            }
        } else {
            mTimes[path] = mTime
        }
    }
    return nil
}

func reloaderLoop(root string) {
    for {
        filepath.Walk(root, visit)
        time.Sleep(500 * time.Millisecond)
    }
}

func start(addr string, port int, path string) error {
    if reloaderEnv := os.Getenv("MANGO_RELOADER"); reloaderEnv != "true" {
        fmt.Printf(" * Knight serving on %s\n", addr)
        stdErrC := make(chan string)
        for {
            fmt.Println(" * Restarting with reloader")
            read, write, _ := os.Pipe()
            go func() {
                var buf bytes.Buffer
                io.Copy(&buf, read)
                stdErrC <- buf.String()
            }()
            arg := []string{"run"}
            _, file := filepath.Split(os.Args[0])
            file = file + ".go"
            arg = append(arg, file)
            arg = append(arg, os.Args[1:]...)
            command := exec.Command("go", arg...)
            command.Env = append(command.Env, "MANGO_RELOADER=true")
            command.Env = append(command.Env, os.Environ()...)
            command.Stdout = os.Stdout
            command.Stderr = write
            err := command.Run()
            if err == nil {
                return nil
            } else {
                write.Close()
                stdErr := <-stdErrC
                if !strings.Contains(stdErr, "exit status 3") {
                    fmt.Print(stdErr)
                    return err
                }
            }
        }
    } else {
        go func() {
            //http.ListenAndServe(addr, nil)
            http.ListenAndServe(fmt.Sprintf("%s:%d", addr, port), nil)
        }()
        reloaderLoop(path)
    }
    return nil
}
