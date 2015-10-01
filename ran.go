package main

import "syscall"
import "os/signal"
import "net/http"
import "os"
import "fmt"
import "strings"
import "sync"
import "github.com/m3ng9i/ran/global"
import "github.com/m3ng9i/ran/server"


func catchSignal() {
    signal_channel := make(chan os.Signal, 1)
    signal.Notify(signal_channel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
    go func() {
        for value := range signal_channel {
            global.Logger.Infof("System: Catch signal: %s, Ran is going to shutdown", value.String())
            global.Logger.Wait()
            os.Exit(0)
        }
    }()
}


func main() {

    global.LoadConfig()

    defer func() {
        global.Logger.Wait()
    }()

    if global.Config.ShowConf {
        for _, line := range strings.Split(global.Config.String(), "\n") {
            line = strings.TrimSpace(line)
            if line != "" {
                global.Logger.Infof("Config: %s", line)
            }
        }
    }

    catchSignal()

    var wg sync.WaitGroup
    defer wg.Wait()

    global.Logger.Infof("System: Ran is running on port %d", global.Config.Port)

    server := server.NewRanServer(global.Config.Config, global.Logger)

    wg.Add(1)
    go func() {
        err := http.ListenAndServe(fmt.Sprintf(":%d", global.Config.Port), server.Serve())
        if err != nil {
            global.Logger.Fatal(err)
        }
        wg.Done()
    }()
}
