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


func startLog() {
    msg := "System: Ran is running on "

    if global.Config.TLS != nil {
        switch global.Config.TLS.Policy {
            case global.TLSRedirect:
                msg += fmt.Sprintf("HTTPS port %d, all traffic from HTTP port %d will redirect to HTTPS port",
                    global.Config.TLS.Port, global.Config.Port)

            case global.TLSBoth:
                msg += fmt.Sprintf("HTTP port %d and HTTPS port %d", global.Config.Port, global.Config.TLS.Port)

            case global.TLSOnly:
                msg += fmt.Sprintf("HTTPS port %d", global.Config.TLS.Port)
        }
    } else {
        msg += fmt.Sprintf("HTTP port %d", global.Config.Port)
    }

    global.Logger.Info(msg)
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

    startLog()

    ran := server.NewRanServer(global.Config.Config, global.Logger)

    startHTTPServer := func() {
        wg.Add(1)
        go func() {
            err := http.ListenAndServe(fmt.Sprintf(":%d", global.Config.Port), ran.Serve())
            if err != nil {
                global.Logger.Fatal(err)
            }
            wg.Done()
        }()
    }

    startTLSServer := func() {
        wg.Add(1)
        go func() {
            err := http.ListenAndServeTLS(fmt.Sprintf(":%d", global.Config.TLS.Port),
                    global.Config.TLS.PublicKey,
                    global.Config.TLS.PrivateKey,
                    ran.Serve())
            if err != nil {
                global.Logger.Fatal(err)
            }
            wg.Done()
        }()
    }

    redirectToHTTPS := func() {
        wg.Add(1)
        go func() {
            err := http.ListenAndServe(fmt.Sprintf(":%d", global.Config.Port), ran.RedirectToHTTPS(global.Config.TLS.Port))
            if err != nil {
                global.Logger.Fatal(err)
            }
            wg.Done()
        }()
    }

    if global.Config.TLS != nil {
        // turn on TLS encryption

        startTLSServer()

        if global.Config.TLS.Policy == global.TLSRedirect {
            redirectToHTTPS()
        } else if global.Config.TLS.Policy == global.TLSBoth {
            startHTTPServer()
        }
    } else {
        startHTTPServer()
    }

}
