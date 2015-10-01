package global

import "fmt"
import "os"
import "github.com/m3ng9i/go-utils/log"


var Logger *log.Logger


func createLogger(debug bool) {
    var config log.Config
    config.Layout       = log.LY_DEFAULT
    config.LayoutStyle  = log.LS_DEFAULT
    config.TimeFormat   = log.TF_DEFAULT
    if debug {
        config.Level = log.DEBUG
    } else {
        config.Level = log.INFO
    }

    var err error
    Logger, err = log.New(os.Stdout, config)
    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
        os.Exit(1)
    }
}

