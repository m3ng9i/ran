package server

import "text/template"
import "fmt"
import "os"
import "time"
import "net/http"
import "net/url"
import "html"
import "path"
import "path/filepath"
import "github.com/dustin/go-humanize"

type dirListFiles struct {
	Name    string
	Url     string
	Size    string
	ModTime time.Time
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
    font-family:"Source Sans Pro",sans-serif;

}

a:link {
    color: #065588;
}

/* visited link */
a:visited {
    color: #2090D8;
}

/* mouse over link */
a:hover {
    color: #0C80CA;
}

/* selected link */
a:active {
    color: #0000FF;
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
    font-size: 15px;
}

table a {
    text-decoration: none;
}

table tr:hover {
    border:1px red solid;
}

table tr > td:nth-child(2), table tr > td:nth-child(3) {
    font-size: 14px;
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
func (this *RanServer) listDir(w http.ResponseWriter, c *context) (size int64, err error) {

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
		// TODO: skipped files according to ignore list in config
		name := i.Name()
		if i.IsDir() {
			name += "/"
		}
		name = html.EscapeString(name)
		url := url.URL{Path: name}

		// write parent dir
		if n == 0 && c.cleanPath != "/" {
			parent := c.parent()

			var info os.FileInfo
			info, err = os.Stat(filepath.Join(this.config.Root, parent))
			if err != nil {
				return
			}

			files = append(files, dirListFiles{Name: "[..]", Url: parent, ModTime: info.ModTime()})
		}

		files = append(files, dirListFiles{Name: name, Url: url.String(), Size: humanize.Bytes(uint64(i.Size())), ModTime: i.ModTime()})
	}

	data := dirList{Title: title, Files: files}

	buf := bufferPool.Get()
	defer bufferPool.Put(buf)

	tplDirList.Execute(buf, data)
	size, _ = buf.WriteTo(w)
	return
}
