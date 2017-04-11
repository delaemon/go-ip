package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"text/template"
	"time"

	"golang.org/x/net/netutil"
)

var (
	d = flag.Bool("d", false, "Whether or not to launch in the background(like a daemon)")
	p = flag.String("p", "8080", "Linten port")
	c = flag.String("c", "100", "Client connection")
)

var (
	accessTime        = time.Now()
	accessLogTemplate = `{{.RemoteAddr}} {{.ContentType}} {{.Method}} {{.Path}} {{.Query}} {{.Body}} {{.UserAgent}}`
)

type accessLogLine struct {
	RemoteAddr  string
	ContentType string
	Path        string
	Query       string
	Method      string
	Body        string
	UserAgent   string
}

func accessLog(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		bufbody := new(bytes.Buffer)
		bufbody.ReadFrom(req.Body)
		body := bufbody.String()
		line := accessLogLine{
			req.RemoteAddr,
			req.Header.Get("Content-Type"),
			req.URL.Path,
			req.URL.RawQuery,
			req.Method, body, req.UserAgent(),
		}
		tmpl, err := template.New("line").Parse(accessLogTemplate)
		if err != nil {
			panic(err)
		}
		bufline := new(bytes.Buffer)
		err = tmpl.Execute(bufline, line)
		if err != nil {
			panic(err)
		}

		logFile := fmt.Sprintf("./log/access/%d%02d%02d.log", accessTime.Year(), accessTime.Month(), accessTime.Day())
		f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic(err)
		}
		log.SetOutput(f)
		log.Printf(bufline.String())

		handler.ServeHTTP(w, req)
	})
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s %s \"%s %s %s %s\" %d %s %s\n",
		r.RemoteAddr,
		r.Header.Get("Content-Type"),
		r.Method,
		r.URL.Path,
		r.Proto,
		r.URL.RawQuery,
		200,
		time.Now(),
		r.UserAgent())
}

func server() {
	cmd := exec.Command(os.Args[0], "-p", *p, "-c", *c)
	serr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}
	err = cmd.Start()
	if err != nil {
		log.Fatalln(err)
	}
	s, err := ioutil.ReadAll(serr)
	s = bytes.TrimSpace(s)
	if bytes.HasPrefix(s, []byte("addr: ")) {
		fmt.Println(string(s))
		cmd.Process.Release()
	} else {
		log.Printf("unexpected response: `%s` error: `%v`\n", s, err)
		cmd.Process.Kill()
	}
}

func client() {
	http.HandleFunc("/", handler)
	listener, err := net.Listen("tcp", ":"+*p)
	if err != nil {
		log.Fatalln(err)
	}
	con, err := strconv.Atoi(*c)
	if err != nil {
		log.Fatalln(err)
	}
	limit_listener := netutil.LimitListener(listener, con)
	http_config := &http.Server{
		Handler: accessLog(http.DefaultServeMux),
	}
	defer limit_listener.Close()
	err = http_config.Serve(limit_listener)
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	flag.Parse()
}

func main() {
	if *d {
		server()
	} else {
		client()
	}
}
