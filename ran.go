package main

import "syscall"
import "os/signal"
import "net"
import "net/http"
import "os"
import "fmt"
import "strings"
import "sync"
import "github.com/m3ng9i/ran/global"
import "github.com/m3ng9i/ran/server"


// version information
var _version_   = "unknown"
var _branch_    = "unknown"
var _commitId_  = "unknown"
var _buildTime_ = "unknown"

var versionInfo = fmt.Sprintf("Version: %s, Branch: %s, Build: %s, Build time: %s",
        _version_, _branch_, _commitId_, _buildTime_)


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



// Get all available IPv4 addresses in system's network interface.
func getIPAutomaticly() (ip []string, e error) {
    iface, err := net.Interfaces()
    if err != nil {
        e = err
        return
    }

    for _, i := range iface {
        addrs, err := i.Addrs()
        if e != nil {
            e = err
            return
        }
        for _, a := range addrs {
            add := net.ParseIP(strings.SplitN(a.String(), "/", 2)[0])
            if add.To4() != nil {
                ip = append(ip, add.String())
            }
        }
    }

    for _, i := range ip {
        if i == "127.0.0.1" {
            return
        }
    }

    // add loopback
    ip = append(ip, "127.0.0.1")

    return
}


// Get all Listening address, like: http://127.0.0.1:8080
func getListeningAddr() (addr []string, err error) {
    ips, err := getIPAutomaticly()
    if err != nil {
        return
    }

    for _, i := range ips {
        if global.Config.TLS != nil {
            if global.Config.TLS.Policy == global.TLSOnly {
                addr = append(addr, fmt.Sprintf("https://%s:%d", i, global.Config.TLS.Port))
            } else {
                addr = append(addr, fmt.Sprintf("http://%s:%d", i, global.Config.Port))
                addr = append(addr, fmt.Sprintf("https://%s:%d", i, global.Config.TLS.Port))
            }
        } else {
            addr = append(addr, fmt.Sprintf("http://%s:%d", i, global.Config.Port))
        }
    }

    return
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

    if global.Config.Auth != nil {
        msg += fmt.Sprintf(" with %s auth", string(global.Config.Auth.Method))
    }

    global.Logger.Info(msg)

    addr, err := getListeningAddr()
    if err != nil {
        global.Logger.Error(err)
    } else {
        for _, i := range addr {
            global.Logger.Infof("System: Listening on %s", i)
        }
    }
}


func main() {

    global.LoadConfig(versionInfo)

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
