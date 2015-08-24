package main
import (
	"fmt"
	"net/http"
	"time"
)

func handler(w http.ResponseWriter, r *http.Request) {
	//addr, err := net.LookupAddr(r.RemoteAddr)
	//if err != nil {
	//	fmt.Println(err)
	//}
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

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
