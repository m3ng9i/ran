package server

import "fmt"
import "errors"
import "net/http"
import "os"
import "time"
import "math/rand"
import "crypto/md5"
import "github.com/m3ng9i/go-utils/log"
import hhelper "github.com/m3ng9i/go-utils/http"


// serveFile() serve any request with content pointed by abspath.
func serveFile(w http.ResponseWriter, r *http.Request, abspath string, setLastModified bool) error {
    f, err := os.Open(abspath)
    if err != nil {
        return err
    }

    info, err := f.Stat()
    if err != nil {
        return err
    }

    if info.IsDir() {
        return errors.New("Cannot serve content of a directory")
    }

    filename := info.Name()

    // TODO if client (use JavaScript) send a request head: 'Accept: "application/octet-stream"' then write the download header ?
    // if the url contains a query like "?download", then download this file
    _, ok := r.URL.Query()["download"]
    if ok {
        hhelper.WriteDownloadHeader(w, filename)
    }

    // if lastModified is not zero Time, http.ServeContent() will write a Last-Modified header.
    var lastModified time.Time
    if setLastModified {
        lastModified = info.ModTime()
    }

    // http.ServeContent() always return a status code of 200.
    http.ServeContent(w, r, filename, lastModified, f)
    return nil
}


type RanServer struct {
    config      Config
    logger      *log.Logger
}


func NewRanServer(c Config, logger *log.Logger) *RanServer {
    return &RanServer {
        config:     c,
        logger:     logger,
    }
}


func setNoCacheHeader(w http.ResponseWriter) {
    w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
    w.Header().Set("Pragma", "no-cache")
    w.Header().Set("Expires", "0")
}


func setCORSHeader(w http.ResponseWriter, r *http.Request) {
    origin := r.Header.Get("Origin")
    if origin == "" {
        origin = "*"
    }
    w.Header().Set("Access-Control-Allow-Origin", origin)
    w.Header().Set("Access-Control-Allow-Credentials", "true")
}


func (this *RanServer) serveHTTP(w http.ResponseWriter, r *http.Request) {

    requestId := string(getRequestId(r.URL.String()))

    w.Header().Set("X-Request-Id", requestId)

    if (this.config.NoCache) {
        setNoCacheHeader(w)
    }

    if (this.config.CORS) {
        setCORSHeader(w, r)
    }

    this.logger.Debugf("#%s: r.URL: [%s]", requestId, r.URL.String())

    context, err := newContext(this.config, r)
    if err != nil {
        Error(w, 500)
        this.logger.Errorf("#%s: %s", requestId, err)
        return
    }

    this.logger.Debugf("#%s: Context: [%s]", requestId, context.String())

    // redirect to a clean url
    if r.URL.String() != context.url {
        http.Redirect(w, r, context.url, http.StatusTemporaryRedirect)
        return
    }

    // display 404 error
    if !context.exist {
        if this.config.Path404 != nil {
            _, err = ErrorFile404(w, *this.config.Path404)
            if err != nil {
                this.logger.Errorf("#%s: Load 404 file error: %s", requestId, err)
                Error(w, 404)
            }
        } else {
            Error(w, 404)
        }
        return
    }

    // display index page
    if context.indexPath != "" {
        err := serveFile(w, r, context.absFilePath, !this.config.NoCache)
        if err != nil {
            Error(w, 500)
            this.logger.Errorf("#%s: %s", requestId, err)
        }
        return
    }

    // display directory list.
    // if c.isDir is true, Config.ListDir must be true,
    // so there is no need to check value of Config.ListDir.
    if context.isDir {
        // display file list of a directory
        _, err = this.listDir(w, this.config.ServeAll, context)
        if err != nil {
            Error(w, 500)
            this.logger.Errorf("#%s: %s", requestId, err)
        }
        return
    }

    // serve the static file.
    err = serveFile(w, r, context.absFilePath, !this.config.NoCache)
    if err != nil {
        Error(w, 500)
        this.logger.Errorf("#%s: %s", requestId, err)
        return
    }

    return
}


// generate a random number in [300,2499], set n for more randomly number
func randTime(n ...int64) int {

    i := time.Now().Unix()
    if len(n) > 0 {
        i += n[0]
    }
    if i < 0 {
        i = 1
    }

    rand.Seed(i)
    return rand.Intn(2200) + 300 // [300,2499]
}


// make the request handler chain:
// log -> authentication -> gzip -> original handler
// TODO: add ip filter: log -> [ip filter] -> authentication -> gzip -> original handler
func (this *RanServer) Serve() http.HandlerFunc {

    // original ran server handler
    handler := this.serveHTTP

    // gzip handler
    if this.config.Gzip {
        handler = hhelper.GzipHandler(handler, true, true)
    }

    // authentication handler
    if this.config.Auth != nil {
        realm := "Identity authentication"

        failFunc := func() {
            // sleep 300~2499 milliseconds to prevent brute force attack
            time.Sleep(time.Duration(randTime()) * time.Millisecond)
        }

        var authFile *hhelper.AuthFile

        // load custom 401 file
        if this.config.Path401 != nil {
            var err error
            authFile, err = errorFile401(this.config)
            if err != nil {
                this.logger.Errorf("Load 401 file error: %s", err)
            }
        }

        if this.config.Auth.Method == DigestMethod {
            da := hhelper.DigestAuth {
                Realm: realm,

                Secret: func(user, realm string) string {
                    if user == this.config.Auth.Username {
                        md5sum := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", user, realm, this.config.Auth.Password)))
                        return fmt.Sprintf("%x", md5sum)
                    }
                    return ""
                },

                ClientCacheSize: 2000,
                ClientCacheTolerance: 200,
            }

            // if authFile is nil, display the default 401 error message
            handler = da.DigestAuthHandler(handler, authFile, failFunc)
        } else {
            ba := hhelper.BasicAuth {
                Realm: realm,
                Secret: hhelper.BasicAuthSecret(this.config.Auth.Username, this.config.Auth.Password),
            }

            handler = ba.BasicAuthHandler(handler, authFile, failFunc)
        }
    }

    // log handler
    handler = this.logHandler(handler)

    return func(w http.ResponseWriter, r *http.Request) {
        handler(w, r)
    }
}


// redirect to https page
func (this *RanServer) RedirectToHTTPS(port uint) http.HandlerFunc {
    handler := this.logHandler(hhelper.RedirectToHTTPS(port))
    return func(w http.ResponseWriter, r *http.Request) {
        requestId := string(getRequestId(r.URL.String()))
        w.Header().Set("X-Request-Id", requestId)
        handler(w, r)
    }
}
