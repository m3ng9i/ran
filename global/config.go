package global

import "os"
import "fmt"
import "flag"
import "strings"
import "path/filepath"
import phelper "github.com/m3ng9i/go-utils/path"
import "github.com/m3ng9i/ran/server"


type TLSPolicy string
const (
    TLSRedirect TLSPolicy = "redirect"
    TLSBoth     TLSPolicy = "both"
    TLSOnly     TLSPolicy = "only"
)


// TLSOption contains options used in TLS encryption
type TLSOption struct {
    PublicKey   string          // Path of public key (certificate)
    PrivateKey  string          // Path of private key
    Port        uint            // HTTPS port. Default is DefaultTLSPort.
    Policy      TLSPolicy       // TLS policy. Default is DefaultTLSPolicy.
}

const DefaultTLSPort uint = 443
const DefaultTLSPolicy = TLSOnly


// Setting about ran server
type Setting struct {
    IP              []string        // IP addresses binded to ran server.
    Port            uint            // HTTP port. Default is 8080.
    ShowConf        bool            // If show config info in the log.
    Debug           bool            // If turns on debug mode. Default is false.
    TLS             *TLSOption      // If is nil, TLS is off.
    errorFile401    *string
    errorFile404    *string
    server.Config
}


// check if path of 404 or 401 file is correct and return an server.ErrorFilePath.
// if the path is not correct, return an error
// p: path of 404 or 401 file, example: /404.html
// name: name of the error file, 401 or 404
func (this *Setting) checkCustomErrorFile(p, name string) (errorFile *server.ErrorFilePath, err error) {
    newPath := filepath.Join(this.Root, p)

    // check if the custom error file is under root
    root := this.Root
    if !strings.HasSuffix(root, string(filepath.Separator)) {
        root = root + string(filepath.Separator)
    }
    if !strings.HasPrefix(newPath, root) {
        err = fmt.Errorf("Path of %s file can not be out of root directory", name)
        return
    }

    // check if the file path is exist and not a directory
    e := phelper.IsExistFile(newPath)
    if e != nil {
        err = fmt.Errorf("'%s': %s", newPath, e)
        return
    }

    relPath, _ := filepath.Rel(root, newPath)

    errorFile = new(server.ErrorFilePath)
    errorFile.Abs = newPath
    errorFile.Rel = "/" + relPath

    return
}


func (this *Setting) check() (errmsg []string) {

    if this.Port > 65535 || this.Port <= 0 {
        errmsg = append(errmsg, "Available HTTP port range is 1-65535")
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
        return
    } else {
        if info.IsDir() == false {
            errmsg = append(errmsg, "Root is not a directory")
            return
        }

        this.Root, err = filepath.Abs(this.Root)
        if err != nil {
            errmsg = append(errmsg, fmt.Sprintf("Can not convert root to absolute form: %s", err.Error()))
            return
        }
    }

    if this.errorFile404 != nil {
        this.Path404, err = this.checkCustomErrorFile(*this.errorFile404, "404")
        if err != nil {
            errmsg = append(errmsg, err.Error())
        }
    }

    if this.Auth != nil {
        if this.Auth.Username == "" || this.Auth.Password == "" {
            errmsg = append(errmsg, "Username or password cannot be empty string")
        }

        if this.Auth.Method != server.BasicMethod && this.Auth.Method != server.DigestMethod {
            errmsg = append(errmsg, "Invalid authentication method")
        }

        for _, p := range this.Auth.Paths {
            if !strings.HasPrefix(p, "/") {
                errmsg = append(errmsg, fmt.Sprintf(`Auth path must start with "/", got %s`, p))
            }
        }

        if this.errorFile401 != nil {
            this.Path401, err = this.checkCustomErrorFile(*this.errorFile401, "401")
            if err != nil {
                errmsg = append(errmsg, err.Error())
            }
        }
    }

    if this.TLS != nil {
        if this.TLS.PublicKey == "" || this.TLS.PrivateKey == "" {
            errmsg = append(errmsg, "Both certificate path and key path should be provided")
        } else {
            if err := phelper.IsNonEmptyFile(this.TLS.PublicKey); err != nil {
                errmsg = append(errmsg, fmt.Sprintf("'%s': %s", this.TLS.PublicKey, err))
            }
            if err := phelper.IsNonEmptyFile(this.TLS.PrivateKey); err != nil {
                errmsg = append(errmsg, fmt.Sprintf("'%s': %s", this.TLS.PrivateKey, err))
            }
        }

        if this.TLS.Port > 65535 || this.TLS.Port <= 0 {
            errmsg = append(errmsg, "Available HTTPS port range is 1-65535")
        }

        if this.TLS.Policy != TLSRedirect && this.TLS.Policy != TLSBoth && this.TLS.Policy != TLSOnly {
            errmsg = append(errmsg, `Value of TLS policy could only be "redirect", "both" or "only"`)
            return // ignore the following checking
        }

        if this.TLS.Policy != TLSOnly && this.TLS.Port == this.Port {
            errmsg = append(errmsg, "HTTP port and HTTPS port cannot be the same.")
        }
    }

    return
}


func (this *Setting) String() string {

https := `TLS: on
Certificate: %s
Private key: %s
TLS port: %d
TLS policy: %s`

    if this.TLS != nil {
        https = fmt.Sprintf(https, this.TLS.PublicKey, this.TLS.PrivateKey, this.TLS.Port, this.TLS.Policy)
    } else {
        https = "TLS: off"
    }

    auth := "<None>"
    if this.Auth != nil {
        auth = string(this.Auth.Method)
    }

s := `Root: %s
Port: %d
Path404: %s
IndexName: %s
ListDir: %t
ServeAll: %t
Gzip: %t
NoCache: %t
CORS: %t
SecureContext: %t
Debug: %t
Auth: %s
Path401: %s
%s`

    path404 := "<None>"
    if this.Path404 != nil {
        path404 = this.Path404.Rel
    }

    path401 := "<None>"
    if this.Path401 != nil {
        path401 = this.Path401.Rel
    }

    s = fmt.Sprintf(s,
                    this.Root,
                    this.Port,
                    path404,
                    strings.Join(this.IndexName, ", "),
                    this.ListDir,
                    this.ServeAll,
                    this.Gzip,
                    this.NoCache,
                    this.CORS,
                    this.SecureContext,
                    this.Debug,
                    auth,
                    path401,
                    https)

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
    c.ServeAll      = false
    c.Gzip          = true
    c.Debug         = false

    return
}


func usage() {
s := `Ran: a simple static web server

Usage: ran [Options...]

Options:

    -r,  -root=<path>           Root path of the site. Default is current working directory.
    -b,  -bind-ip=<ip>          Bind one or more IP addresses to the ran web server.
                                Multiple IP addresses should be separated by comma.
                                If not provide this Option, ran will use 0.0.0.0.
    -p,  -port=<port>           HTTP port. Default is 8080.
         -404=<path>            Path of a custom 404 file, relative to Root. Example: /404.html.
    -i,  -index=<path>          File name of index, priority depends on the order of values.
                                Separate by colon. Example: -i "index.html:index.htm"
                                If not provide, default is index.html and index.htm.
    -l,  -listdir=<bool>        When request a directory and no index file found,
                                if listdir is true, show file list of the directory,
                                if listdir is false, return 404 not found error.
                                Default is false.
    -sa, -serve-all=<bool>      Serve all paths even if the path is start with dot. Default is false.
    -g,  -gzip=<bool>           Turn on or off gzip compression. Default value is true (means turn on).

    -nc, -no-cache=<bool>       If true, ran will remove Last-Modified header and write some no-cache headers to the response:
                                    Cache-Control: no-cache, no-store, must-revalidate
                                    Pragma: no-cache
                                    Expires 0
                                Default is false.

         -cors=<bool>           If true, ran will write some cross-origin resource sharing headers to the response:
                                    Access-Control-Allow-Origin: *
                                    Access-Control-Allow-Credentials: true
                                    Access-Control-Allow-Methods: *
                                    Access-Control-Allow-Headers: *
                                If the request header has a Origin field, then it's value is used in Access-Control-Allow-Origin.
                                Default is false.

         -secure-context=<bool> If true, ran will write some cross-origin security headers to the response:
                                    Cross-Origin-Opener-Policy: same-origin
                                    Cross-Origin-Embedder-Policy: require-corp
                                Default is false.

    -am, -auth-method=<auth>    Set authentication method, valid values are basic and digest. Default is basic.
    -a,  -auth=<user:pass>      Turn on authentication and set username and password (separate by colon).
                                After turn on authentication, all the page require authentication.
         -401=<path>            Path of a custom 401 file, relative to Root. Example: /401.html.
                                If authentication fails and 401 file is set,
                                the file content will be sent to the client.

         -tls-port=<port>       HTTPS port. Default is 443.
         -tls-policy=<pol>      This option indicates how to handle HTTP and HTTPS traffic.
                                There are three option values: redirect, both and only.
                                redirect: redirect HTTP to HTTPS
                                both:     both HTTP and HTTPS are enabled
                                only:     only HTTPS is enabled, HTTP is disabled
                                The default value is: only.
         -cert=<path>           Load a file as a certificate.
                                If use with -make-cert, will generate a certificate to the path.
         -key=<path>            Load a file as a private key.
                                If use with -make-cert, will generate a private key to the path.

Other options:

         -make-cert             Generate a self-signed certificate and a private key used in TLS encryption.
                                You should use -cert and -key to set the output paths.
         -showconf              Show config info in the log.
         -debug                 Turn on debug mode.
    -v,  -version               Show version information.
    -h,  -help                  Show help message.

Author:

    m3ng9i
    <https://github.com/m3ng9i>
    <http://mengqi.info>
`
fmt.Printf(s)
os.Exit(0)
}


func LoadConfig(versionInfo string) {

    var err error
    Config, err = defaultConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, err.Error())
        os.Exit(1)
    }

    var configPath, bindip, root, path404, authMethod, auth, path401, certPath, keyPath, tlsPolicy string
    var port, tlsPort uint
    var indexName server.Index
    var version, help, makeCert bool

    flag.StringVar(&configPath, "c",      "", "Path of config file")
    flag.StringVar(&configPath, "config", "", "Path of config file")

    if configPath != "" {
        // TODO: load config file
    }

    flag.StringVar(&bindip,                 "b",                "",      "IP addresses binded to ran server")
    flag.StringVar(&bindip,                 "bind-ip",          "",      "IP addresses binded to ran server")
    flag.UintVar(  &port,                   "p",                0,       "HTTP port")
    flag.UintVar(  &port,                   "port",             0,       "HTTP port")
    flag.StringVar(&root,                   "r",                "",      "Root path of the website")
    flag.StringVar(&root,                   "root",             "",      "Root path of the website")
    flag.StringVar(&path404,                "404",              "",      "Path of a custom 404 file")
    flag.StringVar(&path401,                "401",              "",      "Path of a custom 401 file")
    flag.StringVar(&authMethod,             "am",               "basic", "authentication method")
    flag.StringVar(&authMethod,             "auth-method",      "basic", "authentication method")
    flag.StringVar(&auth,                   "a",                "",      "Username and password of auth, separate by colon")
    flag.StringVar(&auth,                   "auth",             "",      "Username and password of auth, separate by colon")
    flag.Var(      &indexName,              "i",                         "File name of index, separate by colon")
    flag.Var(      &indexName,              "index",                     "File name of index, separate by colon")
    flag.BoolVar(  &Config.ListDir,         "l",                false,   "Show file list of a directory")
    flag.BoolVar(  &Config.ListDir,         "listdir",          false,   "Show file list of a directory")
    flag.BoolVar(  &Config.ServeAll,        "sa",               false,   "Serve all paths even if the path is start with dot")
    flag.BoolVar(  &Config.ServeAll,        "serve-all",        false,   "Serve all paths even if the path is start with dot")
    flag.BoolVar(  &Config.Gzip,            "g",                true,    "Turn on/off gzip compression")
    flag.BoolVar(  &Config.Gzip,            "gzip",             true,    "Turn on/off gzip compression")
    flag.BoolVar(  &Config.NoCache,         "nc",               false,   "If send no-cache header")
    flag.BoolVar(  &Config.NoCache,         "no-cache",         false,   "If send no-cache header")
    flag.BoolVar(  &Config.CORS,            "cors",             false,   "If send CORS headers")
    flag.BoolVar(  &Config.SecureContext,   "secure-context",   false,   "If send secure context headers")
    flag.BoolVar(  &Config.ShowConf,        "showconf",         false,   "If show config info in the log")
    flag.BoolVar(  &Config.Debug,           "debug",            false,   "Turn on debug mode")
    flag.BoolVar(  &version,                "v",                false,   "Show version information")
    flag.BoolVar(  &version,                "version",          false,   "Show version information")
    flag.BoolVar(  &help,                   "h",                false,   "Show help message")
    flag.BoolVar(  &help,                   "help",             false,   "Show help message")
    flag.BoolVar(  &makeCert,               "make-cert",        false,   "Generate a self-signed certificate and a private key")
    flag.StringVar(&certPath,               "cert",             "",      "Path of certificate")
    flag.StringVar(&keyPath,                "key",              "",      "Path of private key")
    flag.UintVar(  &tlsPort,                "tls-port",         0,       "HTTPS port")
    flag.StringVar(&tlsPolicy,              "tls-policy",       "",      "TLS policy")

    flag.Usage = usage

    flag.Parse()

    if help {
        usage()
    }

    if version {
        fmt.Println(versionInfo)
        os.Exit(0)
    }

    if makeCert {
        err = makeCertFiles(certPath, keyPath, false)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %s\n", err)
            os.Exit(1)
        }
        fmt.Println("Certificate and private key are created")
        os.Exit(0)
    }

    // load TLS config
    if certPath != "" || keyPath != "" || tlsPort > 0 || tlsPolicy != "" {
        if Config.TLS == nil {
            Config.TLS = new(TLSOption)
        }
        Config.TLS.PublicKey    = certPath
        Config.TLS.PrivateKey   = keyPath
        Config.TLS.Port         = tlsPort
        Config.TLS.Policy       = TLSPolicy(tlsPolicy)

    }

    // set default value for Config.TLS
    if Config.TLS != nil {
        if Config.TLS.Port == 0 {
            Config.TLS.Port = DefaultTLSPort
        }
        if Config.TLS.Policy == "" {
            Config.TLS.Policy = DefaultTLSPolicy
        }
    }

    Config.IP, err = getIPs(bindip)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %s\n", err)
        os.Exit(1)
    }

    if port > 0 {
        Config.Port = port
    }

    if root != "" {
        Config.Root = root
    }

    if path404 != "" {
        Config.errorFile404 = &path404
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
            fmt.Fprintln(os.Stderr, "Config error: format of auth not correct")
            os.Exit(1)
        }
        Config.Auth.Username = authPair[0]
        Config.Auth.Password = authPair[1]
        Config.Auth.Method = server.AuthMethod(strings.ToLower(authMethod))

        if path401 != "" {
            Config.errorFile401 = &path401
        }
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


