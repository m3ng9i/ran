package server

import "net/http"
import "fmt"
import "io/ioutil"
import "path"
import hhelper "github.com/m3ng9i/go-utils/http"


// ErrorEx writes http error, accrording status code and msg, return number of bytes write to ResponseWriter.
// Parameter msg is a string contains html, could be ignored.
func ErrorEx(w http.ResponseWriter, code int, title, msg string) int64 {
    status := http.StatusText(code)
    if status == "" {
        status = "Unknown"
    }
    status = fmt.Sprintf("%d %s", code, status)

    if title == "" {
        title = status
    }

    if msg == "" {
        msg = "<h1>" + status + "</h1>"
    }

    html := fmt.Sprintf(`<!DOCTYPE HTML><html><head><meta charset="utf-8"><title>%s</title></head><body>%s</body></html>`,
        title, msg)

    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    w.WriteHeader(code)
    n, _ := fmt.Fprintln(w, html)
    return int64(n)
}


// A convenient way to call ErrorEx()
func Error(w http.ResponseWriter, code int) int64 {
    return ErrorEx(w, code, "", "")
}


// ErrorFile404 writes 404 file to client.
// abspath is path of 404 file.
func ErrorFile404(w http.ResponseWriter, abspath string) (int64, error) {

    b, err := ioutil.ReadFile(abspath)
    if err != nil {
        return 0, err
    }

    contentType, _ := hhelper.FileContentType(path.Ext(abspath))
    if contentType == "" {
        contentType = "text/html; charset=utf-8"
    }
    w.Header().Set("Content-Type", contentType)
    w.WriteHeader(404)
    n, _ := w.Write(b)
    return int64(n), nil
}


func errorFile401(config Config) (a *hhelper.AuthFile, err error) {
    if config.Path401 != nil {
        tp, _ := hhelper.FileContentType(path.Ext(config.Path401.Abs))
        if tp == "" {
            tp = "text/html; charset=utf-8"
        }

        b, e := ioutil.ReadFile(config.Path401.Abs)
        if e != nil {
            err = e
            return
        }

        a = new(hhelper.AuthFile)
        a.ContentType = tp
        a.Body = b

        return
    }

    return
}
