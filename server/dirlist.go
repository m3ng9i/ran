package server

import "html/template"
import "fmt"
import "os"
import "time"
import "net/http"
import "net/url"
import "html"
import "path"
import "path/filepath"
import "strings"
import "sort"


type dirListFiles struct {
    Name        string
    Url         string
    Size        int64
    ModTime     time.Time
}


type dirList struct {
    Title string
    Files []dirListFiles
}


const dirListTpl = `<!DOCTYPE HTML>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="initial-scale=1,width=device-width">
<title>{{.Title}}</title>

<style type="text/css">

body {
    background-color:white;
    color: #333333;
}

table {
    border-collapse: collapse;
}

table tr:nth-child(1) {
    background-color: #f0f0f0;
}

table th, table td {
    padding: 8px 10px;
    border:1px #dddddd solid;
    font-size: 14px;
}

table a {
    text-decoration: none;
}

table tr:hover {
    border:1px red solid;
}

table tr > td:nth-child(2), table tr > td:nth-child(3) {
    font-size: 13px;
}

</style>

</head>

<body>
<h1>{{.Title}}</h1>
<table>
<tr><th>Name</th><th>Size</th><th>Modification time</th></tr>
{{range $files := .Files}}
    <tr>
        <td><a href="{{.Url}}">{{.Name}}</a></td>
        <td>{{.Size}}</td>
        {{/* t2s example: {{ t2s .ModTime "2006-01-02 15:04"}} */}}
        <td>{{t2s .ModTime}}</td>
    </tr>
{{end}}
</table>

</body>
</html>`


var tplDirList *template.Template


func timeToString(t time.Time, format ...string) string {
    f := "2006-01-02 15:04:05"
    if len(format) > 0 && format[0] != "" {
        f = format[0]
    }
    return t.Format(f)
}


func init() {
    var err error
    tplDirList = template.New("dirlist").Funcs(template.FuncMap{"t2s": timeToString})
    tplDirList, err = tplDirList.Parse(dirListTpl)
    if err != nil {
        fmt.Fprintf(os.Stderr, "Directory list template init error: %s", err.Error())
        os.Exit(1)
    }
}


// List content of a directory.
// If error occurs, this function will return an error and won't write anything to ResponseWriter.
func (this *RanServer) listDir(w http.ResponseWriter, serveAll bool, c *context) (size int64, err error) {

    if !c.exist {
        size = Error(w, 404)
        return
    }

    if !c.isDir {
        err = fmt.Errorf("Cannot list contents of a non-directory")
        return
    }

    f, err := os.Open(c.absFilePath)
    if err != nil {
        return
    }
    defer f.Close()

    info, err := f.Readdir(0)
    if err != nil {
        return
    }

    w.Header().Set("Content-Type", "text/html; charset=utf-8")

    title := html.EscapeString(path.Base(c.cleanPath))

    var files []dirListFiles

    for n, i := range info {
        name := i.Name()
        if i.IsDir() {
            name += "/"
        }

        // skip hidden path
        if !serveAll && strings.HasPrefix(name, ".") {
            continue
        }

        fileUrl:= url.URL{Path: name}

        // write parent dir
        if n == 0 && c.cleanPath != "/" {
            parent := c.parent()

            // unescape parent before get it's modification time
            var parentUnescape string
            parentUnescape, err = url.QueryUnescape(parent)
            if err != nil {
                return
            }

            var info os.FileInfo
            info, err = os.Stat(filepath.Join(this.config.Root, parentUnescape))
            if err != nil {
                return
            }

            files = append(files, dirListFiles{Name:"[..]", Url:parent, ModTime:info.ModTime()})
        }

        fileRelPath := path.Join(c.cleanPath, name)

        // skip 404 file
        if this.config.Path404 != nil && fileRelPath == this.config.Path404.Rel {
            continue
        }
        // skip 401 file
        if this.config.Path401 != nil && fileRelPath == this.config.Path401.Rel {
            continue
        }
        files = append(files, dirListFiles{Name:name, Url:fileUrl.String(), Size:i.Size(), ModTime:i.ModTime()})
    }

    sort.Slice(files, func(i, j int) bool {
        return files[i].Name < files[j].Name
    })

    data := dirList{ Title: title, Files: files}

    buf := bufferPool.Get()
    defer bufferPool.Put(buf)

    tplDirList.Execute(buf, data)
    size, _ = buf.WriteTo(w)
    return
}


