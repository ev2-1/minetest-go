package http_media

import (
	"github.com/ev2-1/minetest-go/minetest"

	"net/http"
	"os"
	"strings"
)

type fileHandler struct {
	Root string
}

func (fh *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file := r.URL.RawQuery

	file = strings.ReplaceAll(file, "..", "")

	data, err := os.ReadFile(minetest.Path("hashfiles/" + file))
	if err != nil {
		w.Write([]byte("not found!"))
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

type file struct {
	Path string
}

func (f *file) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile(f.Path)
	if err != nil {
		w.Write([]byte("not found!"))
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(200)
	w.Write(data)
}

func init() {
	fh := &fileHandler{Root: minetest.Path("hashfiles")}

	generateIndex()

	ip := getOutboundIP()

	http.Handle("/mediafile", fh)
	go http.ListenAndServe(":8081", nil)

	// tell minetest where to find:
	minetest.AddMediaURL(ip + ":8081/mediafile?")
}
