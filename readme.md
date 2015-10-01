Ran: a simple static web server written in Go
=============================================

![Ran](ran.gif)

Ran is a simple web server for serving static files.

## Features

- Directory listing
- Automatic gzip compression
- Digest authentication
- Access logging
- Custom 404 error file

## What Ran for?

- File sharing in LAN or home network
- Web application testing
- Personal web site hosting or demonstrating

## Dependencies

- [github.com/abbot/go-http-auth](https://github.com/abbot/go-http-auth)
- [github.com/oxtoacart/bpool](https://github.com/oxtoacart/bpool)
- [github.com/m3ng9i/go-utils/http](https://github.com/m3ng9i/go-utils)
- [github.com/m3ng9i/go-utils/log](https://github.com/m3ng9i/go-utils)
- [github.com/m3ng9i/go-utils/possible](https://github.com/m3ng9i/go-utils)
- [golang.org/x/net/context](https://github.com/golang/net)

## Installation

Use the command below to install the dependencies mentioned above, and build the binary into $GOPATH/bin.

```bash
go get -u github.com/m3ng9i/ran
```

For convenience, you can move the ran binary to a directory in the PATH environment variable.

You can also call `./build.py` command under the Ran source directory to write version information into the binary, so that `ran -v` will give a significant result. Run `./build.py -h` for help.

## Download binary

You can also download Ran binary without build it yourself.

[Download Ran binary from the release page](https://github.com/m3ng9i/ran/releases).

## Run Ran

You can start a web server without any options by typing `ran` and press return in terminal window. This will use the following default configuration:

Configuration               | Default value
----------------------------|--------------------------------
Root directory              | the current working directory
Port                        | 8080
Index file                  | index.html, index.htm
List files of directories   | false
Gzip                        | true
Digest auth                 | false

Open http://127.0.0.1:8080 in browser to see your website.

You can use the options below to override the default configuration.

```
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
```

Example 1: Start a server in the current directory and set port to 8888:

```bash
ran -p=8888
```

Example 2: Set root to /tmp, list files of directories and set a custom 404 page:

```bash
ran -p=/tmp -l=true -404=/404.html
```

Example 3: Close gzip compression, set access username and password:

```bash
ran -g=false -a=user:pass
```

Example 4: Set custom index file:

```bash
ran -i default.html:index.html
```

Other options:

```
        -showconf           Show config info in the log.
        -debug              Turn on debug mode.
    -v, -version            Show version information.
    -h, -help               Show help message.
```

## Tips and tricks

### Execute permission

Before running Ran binary or build.py, make sure they have execute permission. If don't, use `chmod u+x <filename>` to set.

### download parameter

If you add `download` as a query string parameter in the url, the browser will download the file instead of displaying it in the browser window. Example:

```
http://127.0.0.1:8080/readme.html?download
```

### gzip parameter

Gzip compression is enabled by default. Ran will gzip file automaticly according to the file extension. Example: a `.txt` file will be compressed and a `.jpg` file will not.

If you add `gzip=true` in the url, Ran will force compress the file even if the file should not be compressed. Example:

```
http://127.0.0.1:8080/picture.jpg?gzip=true
```

If you add `gzip=false` in the url, Ran will not compress it even if it should be compressed. Example:

```
http://127.0.0.1:8080/large-file.txt?gzip=false
```

Read the source code of [CanBeCompressed()](https://github.com/m3ng9i/go-utils/blob/master/http/can_be_compressed.go) to learn more about automatic gzip compression.

## ToDo

The following functionalities will be added in the future:

- Load config from file
- TLS encryption
- IP filter
- Custom log format
- etc

## What's the meaning of Ran

It's a Chinese PinYin and pronounce 燃, means flame burning.

## Author

mengqi <https://github.com/m3ng9i>

If you like this project, please give me a star.

