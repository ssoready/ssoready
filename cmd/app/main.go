package main

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
)

var srv = flag.String("srv", "", "directory to serve")
var listenAddr = flag.String("listen-addr", "localhost:8083", "listen on address")

func main() {
	flag.Parse()

	srvFS := os.DirFS(*srv)
	fbFS := fallbackFS{
		fsys:     srvFS,
		fallback: "index.html",
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(fbFS)))
	if err := http.ListenAndServe(*listenAddr, logHandler(mux)); err != nil {
		log.Fatal(err)
	}
}

func logHandler(handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println(fmt.Sprintf("%s %s", r.Method, r.URL))
		handler.ServeHTTP(w, r)
	}
}

type fallbackFS struct {
	fsys     fs.FS
	fallback string
}

func (fsys fallbackFS) Open(name string) (fs.File, error) {
	f, err := fsys.fsys.Open(name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fsys.fsys.Open(fsys.fallback)
		}
		return nil, err
	}
	return f, nil
}
