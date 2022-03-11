package server

import "strings"


type Index []string


func (this *Index) String() string {
    return strings.Join([]string(*this), ":")
}


func (this *Index) Set(value string) error {
    *this = Index(strings.Split(value, ":"))
    return nil
}


type AuthMethod string
const BasicMethod  AuthMethod = "basic"
const DigestMethod AuthMethod = "digest"


type Auth struct {
    Username string
    Password string

    // paths which use password to protect, relative to "/".
    // if Paths is empty, all paths are protected.
    // not used currently
    Paths    []string

    Method AuthMethod
}


// ErrorFilePath describe path of a 401/404 file which is under directory of Root.
type ErrorFilePath struct {
    Abs string // Absolute path of error file, e.g. /data/wwwroot/404.html
    Rel string // Path of error file, relative to the root, e.g. /404.html
}


type Config struct {
    Root            string          // Root path of the website. Default is current working directory.
    Path404         *ErrorFilePath  // Path of custom 404 file, under directory of Root.
                                    // When a 404 not found error occurs, the file's content will be send to client.
                                    // nil means do not use 404 file.
    Path401         *ErrorFilePath  // Path of custom 401 file, under directory of Root.
                                    // When a 401 unauthorized error occurs, the file's content will be send to client.
                                    // nil means do not use 401 file.
    IndexName       Index           // File name of index, priority depends on the order of values.
                                    // Default is []string{"index.html", "index.htm"}.
    ListDir         bool            // If no index file provide, show file list of the directory.
                                    // Default is false.
    Gzip            bool            // If turn on gzip compression, default is true.
    NoCache         bool            // If true, ran will write some no-cache headers to the response. Default is false.
    CORS            bool            // If true, ran will write some CORS headers to the response. Default is false.
    SecureContext   bool            // If true, ran will write some cross-origin security headers to the response. Default is false.
    Auth            *Auth           // If not nil, turn on authentication.
    ServeAll        bool            // If is false, path start with dot will not be served, that means a 404 error will be returned.
}


