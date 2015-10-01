package global

import "os"
import "fmt"
import "flag"
import "strings"
import "path/filepath"
import "github.com/m3ng9i/ran/server"


// version information
var _version_   = "unknown"
var _branch_    = "unknown"
var _commitId_  = "unknown"
var _buildTime_ = "unknown"

var versionInfo = fmt.Sprintf("Version: %s, Branch: %s, Build: %s, Build time: %s",
        _version_, _branch_, _commitId_, _buildTime_)


// Setting about ran server
type Setting struct {
    Port        uint            // HTTP port. Default is 8080.
    ShowConf    bool            // If show config info in the log.
    Debug       bool            // If turns on debug mode. Default is false.
    server.Config
}


func (this *Setting) check() (errmsg []string) {

    if this.Port > 65535 || this.Port <= 0 {
        errmsg = append(errmsg, "Available port range is 1-65535")
    }

    for _, index := range this.IndexName {
        name := filepath.Base(index)
        if name != index {
            errmsg = append(errmsg, "Filename of index can not include path separators")
            break
        }
    }

    // If root is not correct, no need to check other variable in Setting structure
    info, err := os.Stat(this.Root)
    if err != nil {
        if os.IsNotExist(err) {
            errmsg = append(errmsg, fmt.Sprintf("Root '%s' is not exist", this.Root))
        } else {
            errmsg = append(errmsg, fmt.Sprintf("Get stat of root directory error: %s", err.Error()))
        }
        goto END
    } else {
        if info.IsDir() == false {
            errmsg = append(errmsg, fmt.Sprintf("Root is not a directory"))
            goto END
        }

        this.Root, err = filepath.Abs(this.Root)
        if err != nil {
            errmsg = append(errmsg, fmt.Sprintf("Can not convert root to absolute form: %s", err.Error()))
            goto END
        }
    }

    if this.Path404 != nil {
        *this.Path404 = filepath.Join(this.Root, *this.Path404)

        // check if 404 file is under root
        root := this.Root
        if !strings.HasSuffix(root, string(filepath.Separator)) {
            root = root + string(filepath.Separator)
        }
        if !strings.HasPrefix(*this.Path404, root) {
            errmsg = append(errmsg, "Path of 404 file can not be out of root directory")
            goto END
        }

        info, err = os.Stat(*this.Path404)
        if err != nil {
            if os.IsNotExist(err) {
                errmsg = append(errmsg, fmt.Sprintf("404 file '%s' is not exist", *this.Path404))
            } else {
                errmsg = append(errmsg, fmt.Sprintf("Get stat of 404 file error: %s", err.Error()))
            }
        } else {
            if info.IsDir() {
                errmsg = append(errmsg, fmt.Sprintf("404 file can not be a directory"))
            }
        }
    }

    if this.Auth != nil {
        if this.Auth.Username == "" || this.Auth.Password == "" {
            errmsg = append(errmsg, "Username or password cannot be empty string")
        }

        for _, p := range this.Auth.Paths {
            if !strings.HasPrefix(p, "/") {
                errmsg = append(errmsg, fmt.Sprintf(`Auth path must start with "/", got %s`, p))
            }
        }
    }

    END: return
}


func (this *Setting) String() string {

s := `Root: %s
Port: %d
Path404: %s
IndexName: %s
ListDir: %t
Gzip: %t
Debug: %t
Digest auth: %t`

    path404 := "<None>"
    if this.Path404 != nil {
        path404 = *this.Path404
    }

    s = fmt.Sprintf(s,
                    this.Root,
                    this.Port,
                    path404,
                    strings.Join(this.IndexName, ", "),
                    this.ListDir,
                    this.Gzip,
                    this.Debug,
                    !(this.Auth == nil))

    return s
}


var Config *Setting


func defaultConfig() (c *Setting, err error) {
    c = new(Setting)

    c.Root, err = os.Getwd()
    if err != nil {
        return
    }

    c.Port          = 8080
    c.Path404       = nil
    c.IndexName     = []string{"index.html", "index.htm"}
    c.ListDir       = false
    c.Gzip          = true
    c.Debug         = false

    return
}


func usage() {
s := `Ran: a simple static web server

Usage: ran [Options...]

Options:

    -r, -root=<path>        Root path of the site. Default is current working directory.
    -p, -port=<port>        HTTP port. Default is 8080.
        -404=<path>         Path of a custom 404 file, relative to Root. Example: /404.html.
    -i, -index=<path>       File name of index, priority depends on the order of values.
                            Separate by colon. Example: -i "index.html:index.htm"
                            If not provide, default is index.html and index.htm.
    -l, -listdir=<bool>     When request a directory and no index file found,
                            if listdir is true, show file list of the directory,
                            if listdir is false, return 404 not found error.
                            Default is false.
    -g, -gzip=<bool>        If turn on gzip compression. Default is true.
    -a, -auth=<user:pass>   Turn on digest auth and set username and password (separate by colon).
                            After turn on digest auth, all the page require authentication.

Other options:

        -showconf           Show config info in the log.
        -debug              Turn on debug mode.
    -v, -version            Show version information.
    -h, -help               Show help message.

Author:

    m3ng9i
    <https://github.com/m3ng9i>
    <http://mengqi.info>
`
fmt.Printf(s)
os.Exit(0)
}


func LoadConfig() {

    var err error
    Config, err = defaultConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
        os.Exit(1)
    }

    var configPath, root, path404, auth string
    var port uint
    var indexName server.Index
    var version, help bool

    flag.StringVar(&configPath, "c",      "", "Path of config file")
    flag.StringVar(&configPath, "config", "", "Path of config file")

    if configPath != "" {
        // TODO: load config file
    }

    flag.UintVar(  &port,            "p",        0,     "HTTP port")
    flag.UintVar(  &port,            "port",     0,     "HTTP port")
    flag.StringVar(&root,            "r",        "",    "Root path of the website")
    flag.StringVar(&root,            "root",     "",    "Root path of the website")
    flag.StringVar(&path404,         "404",      "",    "Path of a custom 404 file")
    flag.StringVar(&auth,            "a",        "",    "Username and password of digest auth, separate by colon")
    flag.StringVar(&auth,            "auth",     "",    "Username and password of digest auth, separate by colon")
    flag.Var(      &indexName,       "i",               "File name of index, separate by colon")
    flag.Var(      &indexName,       "index",           "File name of index, separate by colon")
    flag.BoolVar(  &Config.ListDir,  "l",        false, "Show file list of a directory")
    flag.BoolVar(  &Config.ListDir,  "listdir",  false, "Show file list of a directory")
    flag.BoolVar(  &Config.Gzip,     "g",        true,  "Turn on/off gzip compression")
    flag.BoolVar(  &Config.Gzip,     "gzip",     true,  "Turn on/off gzip compression")
    flag.BoolVar(  &Config.ShowConf, "showconf", false, "If show config info in the log")
    flag.BoolVar(  &Config.Debug,    "debug",    false, "Turn on debug mode")
    flag.BoolVar(  &version,         "v",        false, "Show version information")
    flag.BoolVar(  &version,         "version",  false, "Show version information")
    flag.BoolVar(  &help,            "h",        false, "-h")
    flag.BoolVar(  &help,            "help",     false, "-help")

    flag.Usage = usage

    flag.Parse()

    if help {
        usage()
    }

    if version {
        fmt.Println(versionInfo)
        os.Exit(0)
    }

    if port > 0 {
        Config.Port = port
    }

    if root != "" {
        Config.Root = root
    }

    if path404 != "" {
        Config.Path404 = &path404
    }

    if len(indexName) > 0 {
        Config.IndexName = indexName
    }

    if auth != "" {
        if Config.Auth == nil {
            Config.Auth = new(server.Auth)
        }
        authPair := strings.SplitN(auth, ":", 2)
        if len(authPair) != 2 {
            fmt.Fprintf(os.Stderr, "Config error: format of auth not correct")
            os.Exit(1)
        }
        Config.Auth.Username = authPair[0]
        Config.Auth.Password = authPair[1]
    }

    // check Config
    errmsg := Config.check()
    if len(errmsg) == 1 {
        fmt.Fprintf(os.Stderr, "Config error: %s\n", errmsg[0])
        os.Exit(1)
    } else if len(errmsg) > 1 {
        fmt.Fprintln(os.Stderr, "Config error:")
        for i, msg := range errmsg {
            fmt.Fprintf(os.Stderr, "%d. %s\n", i + 1, msg)
        }
        os.Exit(1)
    }

    createLogger(Config.Debug)

    return
}

