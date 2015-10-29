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


type Auth struct {
    Username string
    Password string
    Paths    []string   // paths which use password to protect, relative to "/"
                        // if Paths is empty, all paths are protected.
}


type Config struct {
    Root        string          // Root path of the website. Default is current working directory.
    Path404     *string         // Abspath of custom 404 file, under directory of Root.
                                // When a 404 not found error occurs, the file's content will be send to client.
                                // nil means do not use 404 file.
    Path401     *string         // Abspath of custom 401 file, under directory of Root.
                                // When a 401 unauthorized error occurs, the file's content will be send to client.
                                // nil means do not use 401 file.
    IndexName   Index           // File name of index, priority depends on the order of values.
                                // Default is []string{"index.html", "index.htm"}.
    ListDir     bool            // If no index file provide, show file list of the directory.
                                // Default is false.
    Gzip        bool            // If turn on gzip compression, default is true.
    Auth        *Auth           // If not nil, turn on digest auth.
}


