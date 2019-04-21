package server

import "strconv"
import "errors"
import "fmt"
import "strings"
import "net/http"
import "time"
import hhelper "github.com/m3ng9i/go-utils/http"


type Header http.Header


func (this Header) String() string {
    var s []string
    for key, value := range this {
        for _, val := range value {
            s = append(s, fmt.Sprintf("%s: %s", key, val))
        }
    }
    return strings.Join(s, ", ")
}


var ErrInvalidLogLayout = errors.New("Invalid log layout")

/*
LogLayout indicate what information will be present in access log. LogLayout is a string contains format specifiers. The format specifiers is a tag start with a percent sign and followd by a letter. The format specifiers will be replaced by corresponding values when the log is created.

Below are format specifiers and there meanings:

%%  Percent sign (%)
%i  Request id
%s  Response status code
%h  Host
%a  Client ip address
%m  Request method
%l  Request url
%r  Referer
%u  User agent
%n  Number of bytes transferred
%t  Response time
%c  Compression status (gzip / none)
%S  Scheme (http or https)
*/
type LogLayout string


var LogLayoutNormal LogLayout = `Access #%i: [Status: %s] [Host: %h] [IP: %a] [Method: %m] [Scheme: %S] [URL: %l] [Referer: %r] [UA: %u] [Size: %n] [Time: %t] [Compression: %c]`


var LogLayoutShort LogLayout = `Access #%i: [%s] [%h] [%a] [%m] [%S] [%l] [%r] [%u] [%n] [%t] [%c]`


var LogLayoutMin LogLayout = `Access #%i: [%s] [%a] [%m] [%l] [%n]`


// IsLegal checks if a log layout is legal.
func (this *LogLayout) IsLegal() bool {

    var in bool

    OUTER:
    for _, c := range *this {
        if in {
            for _, ch := range []rune("%ishamlruntcS") {
                if c == ch {
                    in = false
                    continue OUTER
                }
            }
            return false
        } else {
            if c == '%' {
                in = true
            }
        }
    }

    return true
}


func (this *RanServer) accessLog(sniffer *hhelper.ResponseSniffer, r *http.Request, responseTime int64) error {

    buf := bufferPool.Get()
    defer bufferPool.Put(buf)

    var in bool
    // TODO read layout from config
    for _, c := range LogLayoutNormal {
        if in {
            switch c {
                case '%':
                    buf.WriteString("%")

                // request id
                case 'i':
                    buf.WriteString(sniffer.Header().Get("X-Request-Id"))

                // response status code
                case 's':
                    buf.WriteString(strconv.Itoa(sniffer.Code))

                // host
                case 'h':
                    buf.WriteString(r.Host)

                // client ip address
                case 'a':
                    ip := hhelper.GetIP(r)
                    realIp := r.Header.Get("X-Real-Ip")
                    if realIp != "" {
                        ip = ip + " (X-REAL-IP: " + realIp + ")"
                    }
                    buf.WriteString(ip)

                // request method
                case 'm':
                    buf.WriteString(r.Method)

                // request url
                case 'l':
                    buf.WriteString(r.URL.String())

                // referer
                case 'r':
                    buf.WriteString(r.Referer())

                // user agent
                case 'u':
                    buf.WriteString(r.Header.Get("User-Agent"))

                // number of bytes transferred
                case 'n':
                    buf.WriteString(strconv.Itoa(sniffer.Size))

                // response time
                case 't':
                    rt := float64(responseTime) / 1000000
                    buf.WriteString(fmt.Sprintf("%.3fms", rt))

                // compression status (gzip / none)
                case 'c':
                    contentEncoding := strings.ToLower(sniffer.Header().Get("Content-Encoding"))
                    if strings.Contains(contentEncoding, "gzip") {
                        buf.WriteString("gzip")
                    } else {
                        buf.WriteString("none")
                    }

                // scheme
                case 'S':
                    // Because r.URL.Scheme from the request is always empty,
                    // so it's need to use r.TLS to check the scheme.
                    if r.TLS != nil {
                        buf.WriteString("https")
                    } else {
                        buf.WriteString("http")
                    }

                default:
                    return ErrInvalidLogLayout
            }

            in = false
        } else {
            if c == '%' {
                in = true
            } else {
                buf.WriteRune(c)
            }
        }

    }

    this.logger.Info(buf.String())
    return nil
}


func (this *RanServer) logHandler(fn http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        startTime := time.Now()

        sniffer := hhelper.NewSniffer(w, false)

        fn(sniffer, r)

        requestId := sniffer.Header().Get("X-Request-Id")

        this.logger.Debugf("#%s: Response headers: [%s]", requestId, Header(sniffer.Header()).String())

        responseTime := time.Since(startTime).Nanoseconds()

        err := this.accessLog(sniffer, r, responseTime)
        if err != nil {
            this.logger.Errorf("#%s: accessLog(): %s", requestId, err)
        }
    }
}


