package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var (
	d = flag.Bool("d", false, "Whether or not to launch in the background(like a daemon)")
	p = flag.String("p", "8080", "Linten port")
)

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

func init() {
	flag.Parse()
}

func main() {
	if *d {
		cmd := exec.Command(os.Args[0], "-p", *p)
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
	} else {
		http.HandleFunc("/", handler)
		http.ListenAndServe(":"+*p, nil)
	}
}
